package pipeline

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/schema"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/configs"
	"github.com/zclconf/go-cty/cty"
)

// ToError formats the supplied value as an error (or just returns it if already an error)
func ToError(val interface{}) error {
	if e, ok := val.(error); ok {
		return e
	} else {
		// return fperr.InternalWithMessage(fmt.Sprintf("%v", val))
		return fmt.Errorf("%v", val)
	}
}

func LoadFlowpipeConfig(ctx context.Context, configPath string) (*FlowpipeConfigParseContext, error) {
	parseCtx := NewFlowpipeConfigParseContext(ctx, configPath)

	// check whether sourcePath is a glob with a root location which exists in the file system
	localSourcePath, globPattern, err := filehelpers.GlobRoot(configPath)
	if err != nil {
		return nil, err
	}

	if localSourcePath == globPattern {
		// if the path is a folder,
		// append '*' to the glob explicitly, to match all files in that folder.
		globPattern = path.Join(globPattern, fmt.Sprintf("*%s", pipeparser.PipelineExtension))
	}

	flowpipeConfigFilePaths, err := filehelpers.ListFiles(localSourcePath, &filehelpers.ListOptions{
		Flags:   filehelpers.AllRecursive,
		Include: []string{globPattern},
	})
	if err != nil {
		return nil, err
	}

	// pipelineFilePaths is the list of all pipeline files found in the pipelinePath
	if len(flowpipeConfigFilePaths) == 0 {
		return parseCtx, nil
	}

	fileData, diags := pipeparser.LoadFileData(flowpipeConfigFilePaths...)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	if len(fileData) != len(flowpipeConfigFilePaths) {
		// return nil, fperr.InternalWithMessage("Failed to load all pipeline files")
		return nil, fmt.Errorf("Failed to load all pipeline files")
	}

	// Each file in the pipelineFilePaths is parsed and the result is stored in the bodies variable
	// bodies.data length should be the same with pipelineFilePaths length
	bodies, diags := pipeparser.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := bodies.Content(FlowpipeConfigBlockSchema)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	err = parseAllFlowipeConfig(parseCtx)
	if err != nil {
		return parseCtx, err
	}

	return parseCtx, nil
}

func LoadPipelines(ctx context.Context, configPath string) (map[string]*Pipeline, error) {
	fpParseContext, err := LoadFlowpipeConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}
	return fpParseContext.PipelineHcls, nil
}

func parseAllFlowipeConfig(parseCtx *FlowpipeConfigParseContext) error {
	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags := decodeFlowpipeConfigBlocks(parseCtx)
		if diags.HasErrors() {
			// Store the diagnostics in the parse context, useful for test and introspection
			parseCtx.ParseContext.Diags = diags
			return pipeparser.DiagsToError("Failed to decode pipelines", diags)
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
			_, res := decodePipeline(block, parseCtx)
			diags = append(diags, res.Diags...)
		}
	}

	parseCtx.BuildEvalContext()

	for _, block := range blocksToDecode {
		switch block.Type {
		case schema.BlockTypeTrigger:
			_, res := decodeTrigger(block, parseCtx)
			diags = append(diags, res.Diags...)
		}
	}
	return diags
}

func decodeTrigger(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (ITrigger, *pipeparser.DecodeResult) {
	res := pipeparser.NewDecodeResult()

	if len(block.Labels) != 2 {
		res.HandleDecodeDiags(hcl.Diagnostics{
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
		res.HandleDecodeDiags(hcl.Diagnostics{
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
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	triggerHcl := NewTrigger(parseCtx.RunCtx, triggerType, triggerName)

	if triggerHcl == nil {
		res.HandleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid trigger type '%s'", triggerType),
				Subject:  &block.DefRange,
			},
		})
		return nil, res
	}

	diags = triggerHcl.SetAttributes(triggerOptions.Attributes, &parseCtx.ParseContext)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	moreDiags := parseCtx.AddTrigger(triggerHcl)
	res.AddDiags(moreDiags)

	return triggerHcl, res
}

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*Pipeline, *pipeparser.DecodeResult) {
	res := pipeparser.NewDecodeResult()

	// get shell pipelineHcl
	pipelineHcl := NewPipelineHcl(block)

	// do a partial decode so we can parse the step manually, each pipeline step has its own struct, so we can't use
	// HCL automatic parsing here
	pipelineOptions, rest, diags := block.Body.PartialContent(PipelineBlockSchema)
	if diags.HasErrors() {
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, pipelineHcl)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	diags = pipelineHcl.SetAttributes(pipelineOptions.Attributes)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
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
				res.HandleDecodeDiags(diags)
				return nil, res
			}

			pipelineHcl.Steps = append(pipelineHcl.Steps, step)

		case schema.BlockTypePipelineOutput:
			output, cfgDiags := decodeOutput(block, parseCtx)
			diags = append(diags, cfgDiags...)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
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
				res.HandleDecodeDiags(diags)
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
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	return pipelineHcl, res
}

func validatePipelineDependencies(pipelineHcl *Pipeline) hcl.Diagnostics {
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
					Summary:  fmt.Sprintf("invalid depends_on '%s' - step '%s' does not exist for pipeline %s", dep, step.GetFullyQualifiedName(), pipelineHcl.Name),
				})
			}
		}
	}

	return diags
}

func handlePipelineDecodeResult(resource *Pipeline, res *pipeparser.DecodeResult, block *hcl.Block, parseCtx *FlowpipeConfigParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded()
		res.AddDiags(moreDiags)

		moreDiags = parseCtx.AddPipeline(resource)
		res.AddDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		// moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		moreDiags := parseCtx.AddDependencies(block, resource.Name, res.Depends)
		res.AddDiags(moreDiags)
	}
}

func GetPipelineStepBlockSchema(stepType string) *hcl.BodySchema {
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return PipelineStepHttpBlockSchema
	case schema.BlockTypePipelineStepSleep:
		return PipelineStepSleepBlockSchema
	case schema.BlockTypePipelineStepEmail:
		return PipelineStepEmailBlockSchema
	case schema.BlockTypePipelineStepEcho:
		return PipelineStepEchoBlockSchema
	case schema.BlockTypePipelineStepQuery:
		return PipelineStepQueryBlockSchema
	case schema.BlockTypePipelineStepPipeline:
		return PipelineStepPipelineBlockSchema
	default:
		return nil
	}
}

func GetTriggerBlockSchema(triggerType string) *hcl.BodySchema {
	switch triggerType {
	case schema.TriggerTypeSchedule:
		return TriggerScheduleBlockSchema
	case schema.TriggerTypeInterval:
		return TriggerIntervalBlockSchema
	case schema.TriggerTypeQuery:
		return TriggerQueryBlockSchema
	case schema.TriggerTypeHttp:
		return TriggerHttpBlockSchema
	default:
		return nil
	}
}

type FlowpipeConfigParseContext struct {
	parse.ParseContext
	PipelineHcls map[string]*Pipeline
	TriggerHcls  map[string]ITrigger
}

func (c *FlowpipeConfigParseContext) BuildEvalContext() {
	vars := map[string]cty.Value{}
	pipelineVars := map[string]cty.Value{}

	for _, pipeline := range c.PipelineHcls {
		pipelineVars[pipeline.Name] = pipeline.AsCtyValue()
	}

	vars["pipeline"] = cty.ObjectVal(pipelineVars)

	c.ParseContext.BuildEvalContext(vars)
}

// AddPipeline stores this resource as a variable to be added to the eval context. It alse
func (c *FlowpipeConfigParseContext) AddPipeline(pipelineHcl *Pipeline) hcl.Diagnostics {
	c.PipelineHcls[pipelineHcl.Name] = pipelineHcl

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, pipelineHcl.Name)

	c.BuildEvalContext()
	return nil
}

func (c *FlowpipeConfigParseContext) AddTrigger(trigger ITrigger) hcl.Diagnostics {

	c.TriggerHcls[trigger.GetName()] = trigger

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, trigger.GetName())

	c.BuildEvalContext()
	return nil
}

func NewFlowpipeConfigParseContext(ctx context.Context, rootEvalPath string) *FlowpipeConfigParseContext {
	parseContext := parse.NewParseContext(ctx, rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done

	c := &FlowpipeConfigParseContext{
		ParseContext: parseContext,
		PipelineHcls: make(map[string]*Pipeline),
		TriggerHcls:  make(map[string]ITrigger),
	}

	c.BuildEvalContext()

	return c
}
