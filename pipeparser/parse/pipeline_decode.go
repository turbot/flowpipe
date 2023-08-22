package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/configs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func decodeStep(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (modconfig.IPipelineStep, hcl.Diagnostics) {
	stepType := block.Labels[0]
	stepName := block.Labels[1]

	// TODO: collect all diags?

	step := modconfig.NewPipelineStep(stepType, stepName)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
			Subject:  &block.DefRange,
		}}
	}

	pipelineStepBlockSchema := GetPipelineStepBlockSchema(stepType)
	if pipelineStepBlockSchema == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Pipeline step block schema not found for step " + stepType,
			Subject:  &block.DefRange,
		}}
	}

	stepOptions, rest, diags := block.Body.PartialContent(pipelineStepBlockSchema)

	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, step)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = step.SetAttributes(stepOptions.Attributes, parseCtx.ParseContext.EvalCtx)
	if len(diags) > 0 {
		return nil, diags
	}

	if errorBlocks := stepOptions.Blocks.ByType()[schema.BlockTypeError]; len(errorBlocks) > 0 {
		if len(errorBlocks) > 1 {
			return nil, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Multiple error blocks found for step " + stepName,
				Subject:  &block.DefRange,
			}}
		}
		errorBlock := errorBlocks[0]

		attributes, diags := errorBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return nil, diags
		}

		ignore := false
		retries := 0

		if attr, exists := attributes[schema.AttributeTypeIgnore]; exists {
			val, diags := attr.Expr.Value(nil)
			if len(diags) > 0 {
				return nil, diags
			}

			var target bool
			if err := gocty.FromCtyValue(val, &target); err != nil {
				return nil, hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error decoding ignore attribute",
					Detail:   err.Error(),
					Subject:  &block.DefRange,
				}}
			}
			ignore = target
		}

		if attr, exists := attributes[schema.AttributeTypeRetries]; exists {
			val, diags := attr.Expr.Value(nil)
			if len(diags) > 0 {
				return nil, diags
			}

			var target int
			if err := gocty.FromCtyValue(val, &target); err != nil {
				return nil, hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error decoding retries attribute",
					Detail:   err.Error(),
					Subject:  &block.DefRange,
				}}
			}
			retries = target

		}

		errorConfig := &modconfig.ErrorConfig{
			Ignore:  ignore,
			Retries: retries,
		}

		step.SetErrorConfig(errorConfig)
	} else {
		errorConfig := &modconfig.ErrorConfig{
			Ignore:  false,
			Retries: 0,
		}
		step.SetErrorConfig(errorConfig)
	}

	return step, hcl.Diagnostics{}
}

func decodeOutput(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*modconfig.PipelineOutput, hcl.Diagnostics) {

	o := &modconfig.PipelineOutput{
		Name: block.Labels[0],
	}

	outputOptions, rest, diags := block.Body.PartialContent(modconfig.PipelineOutputBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, outputOptions)
	if len(diags) > 0 {
		return nil, diags
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeSensitive]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Sensitive)
		diags = append(diags, valDiags...)
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeValue]; exists {
		expr := attr.Expr
		if len(expr.Variables()) > 0 {
			traversals := expr.Variables()
			for _, traversal := range traversals {
				parts := hclhelpers.TraversalAsStringSlice(traversal)
				if len(parts) > 0 {
					if parts[0] == schema.BlockTypePipelineStep {
						dependsOn := parts[1] + "." + parts[2]
						o.AppendDependsOn(dependsOn)
					}
				}
			}
		}
		o.UnresolvedValue = attr.Expr

	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing value attribute",
			Subject:  &block.DefRange,
		})
	}

	return o, diags
}

func ParseAllFlowipeConfig(parseCtx *FlowpipeConfigParseContext) error {
	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags := decodeFlowpipeConfigBlocks(parseCtx)
		if diags.HasErrors() {
			// Store the diagnostics in the parse context, useful for test and introspection
			parseCtx.ParseContext.Diags = diags
			return error_helpers.HclDiagsToError("Failed to decode pipelines", diags)
		}

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			// return fperr.BadRequestWithMessage("failed to resolve dependencies after " + fmt.Sprintf("%d", attempts+1) + " passes: " + str)
			return fmt.Errorf("failed to resolve dependencies after %d passes: %s", attempts+1, str)
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}
	return nil

}

func decodeFlowpipeConfigBlocks(parseCtx *FlowpipeConfigParseContext) hcl.Diagnostics {

	var diags hcl.Diagnostics
	blocksToDecode, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	// First decode all the pipelines, then decode the triggers
	//
	// Triggers reference pipelines so it's an easy way to ensure that we have
	// all the pipelines loaded in memory before we try to parse the triggers.
	//
	// TODO: may need to change the logic later when we start validating mod dependencies
	// TODO: because that needs to be done in a higher level (?)
	for _, block := range blocksToDecode {
		switch block.Type {
		case schema.BlockTypePipeline:
			pipelineHcl, res := decodePipeline(nil, block, parseCtx)
			diags = append(diags, res.Diags...)

			if pipelineHcl != nil {
				moreDiags := parseCtx.AddPipeline(pipelineHcl)
				res.addDiags(moreDiags)
			}
		}
	}

	parseCtx.BuildEvalContext()

	for _, block := range blocksToDecode {
		switch block.Type {
		case schema.BlockTypeTrigger:
			// TODO: fix the nil mod passing here
			triggerHcl, res := decodeTrigger(nil, block, parseCtx)
			diags = append(diags, res.Diags...)
			if triggerHcl != nil {
				moreDiags := parseCtx.AddTrigger(triggerHcl)
				res.addDiags(moreDiags)
			}
		}
	}
	return diags
}

func decodeTrigger(mod *modconfig.Mod, block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*modconfig.Trigger, *DecodeResult) {
	res := newDecodeResult()

	if len(block.Labels) != 2 {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	triggerType := block.Labels[0]
	triggerName := block.Labels[1]

	triggerSchema := GetTriggerBlockSchema(triggerType)
	if triggerSchema == nil {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger type '%s'", triggerType),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	triggerOptions, _, diags := block.Body.PartialContent(triggerSchema)

	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	triggerHcl := modconfig.NewTrigger(parseCtx.RunCtx, mod, triggerType, triggerName)

	if triggerHcl == nil {
		res.handleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger type '%s'", triggerType),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	// TODO: wrong location, we will keep rebuilding the Eval Context for every single trigger?
	evalContext := rebuildEvalContextWithCurrentMod(mod, parseCtx.EvalCtx)
	diags = triggerHcl.Config.SetAttributes(mod, triggerHcl, triggerOptions.Attributes, evalContext)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	moreDiags := parseCtx.AddTrigger(triggerHcl)
	res.addDiags(moreDiags)

	return triggerHcl, res
}

func rebuildEvalContextWithCurrentMod(mod *modconfig.Mod, evalContext *hcl.EvalContext) *hcl.EvalContext {

	modName := "local"

	if mod != nil {
		modName = mod.Name()
		parts := strings.Split(modName, ".")
		if len(parts) == 2 {
			modName = parts[1]
		}
	}
	// pulls the current mod data from the eval context
	curentModVars := evalContext.Variables[modName]
	if curentModVars == cty.NilVal {
		return evalContext
	}

	currentModVarsMap := curentModVars.AsValueMap()
	if currentModVarsMap == nil {
		return evalContext
	}

	for k, v := range currentModVarsMap {
		evalContext.Variables[k] = v
	}

	return evalContext
}

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(mod *modconfig.Mod, block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*modconfig.Pipeline, *DecodeResult) {
	res := newDecodeResult()

	// get shell pipelineHcl
	pipelineHcl := modconfig.NewPipelineHcl(mod, block)

	// do a partial decode so we can parse the step manually, each pipeline step has its own struct, so we can't use
	// HCL automatic parsing here
	pipelineOptions, rest, diags := block.Body.PartialContent(modconfig.PipelineBlockSchema)
	if diags.HasErrors() {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, pipelineHcl)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	diags = pipelineHcl.SetAttributes(pipelineOptions.Attributes)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	// TODO: should we return immediately after error?

	// use a map keyed by a string for fast lookup
	// we use an empty struct as the value type, so that
	// we don't use up unnecessary memory
	// foundOptions := map[string]struct{}{}
	for _, block := range pipelineOptions.Blocks {
		switch block.Type {
		case schema.BlockTypePipelineStep:
			step, diags := decodeStep(block, parseCtx)
			if diags.HasErrors() {
				res.handleDecodeDiags(diags)
				return nil, res
			}

			pipelineHcl.Steps = append(pipelineHcl.Steps, step)

		case schema.BlockTypePipelineOutput:
			output, cfgDiags := decodeOutput(block, parseCtx)
			diags = append(diags, cfgDiags...)
			if len(diags) > 0 {
				res.handleDecodeDiags(diags)
				return nil, res
			}

			if output != nil {
				pipelineHcl.Outputs = append(pipelineHcl.Outputs, *output)
			}

		case schema.BlockTypeParam:
			override := false
			param, varDiags := configs.DecodeVariableBlock(block, override)
			diags = append(diags, varDiags...)
			if len(diags) > 0 {
				res.handleDecodeDiags(diags)
				return nil, res
			}

			if param != nil {
				pipelineHcl.Params[param.Name] = param
			}

		default:
			// this should never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for workspace profiles", block.Type),
				Subject:  &block.DefRange,
			})
		}
	}

	handlePipelineDecodeResult(pipelineHcl, res, block, parseCtx)
	diags = validatePipelineDependencies(pipelineHcl)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
		return nil, res
	}

	return pipelineHcl, res
}

func validatePipelineDependencies(pipelineHcl *modconfig.Pipeline) hcl.Diagnostics {
	var diags hcl.Diagnostics

	var stepRegisters []string
	for _, step := range pipelineHcl.Steps {
		stepRegisters = append(stepRegisters, step.GetFullyQualifiedName())
	}

	for _, step := range pipelineHcl.Steps {
		dependsOn := step.GetDependsOn()

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s' - step '%s' does not exist for pipeline %s", dep, step.GetFullyQualifiedName(), pipelineHcl.Name()),
				})
			}
		}
	}

	return diags
}

func handlePipelineDecodeResult(resource *modconfig.Pipeline, res *DecodeResult, block *hcl.Block, parseCtx *FlowpipeConfigParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		// moreDiags := resource.OnDecoded()
		// res.addDiags(moreDiags)

		moreDiags := parseCtx.AddPipeline(resource)
		res.addDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.addDiags(moreDiags)
	}
}

func GetPipelineStepBlockSchema(stepType string) *hcl.BodySchema {
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return modconfig.PipelineStepHttpBlockSchema
	case schema.BlockTypePipelineStepSleep:
		return modconfig.PipelineStepSleepBlockSchema
	case schema.BlockTypePipelineStepEmail:
		return modconfig.PipelineStepEmailBlockSchema
	case schema.BlockTypePipelineStepEcho:
		return modconfig.PipelineStepEchoBlockSchema
	case schema.BlockTypePipelineStepQuery:
		return modconfig.PipelineStepQueryBlockSchema
	case schema.BlockTypePipelineStepPipeline:
		return modconfig.PipelineStepPipelineBlockSchema
	default:
		return nil
	}
}

func GetTriggerBlockSchema(triggerType string) *hcl.BodySchema {
	switch triggerType {
	case schema.TriggerTypeSchedule:
		return modconfig.TriggerScheduleBlockSchema
	case schema.TriggerTypeInterval:
		return modconfig.TriggerIntervalBlockSchema
	case schema.TriggerTypeQuery:
		return modconfig.TriggerQueryBlockSchema
	case schema.TriggerTypeHttp:
		return modconfig.TriggerHttpBlockSchema
	default:
		return nil
	}
}
