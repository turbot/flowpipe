package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/pipe-fittings/schema"
)

// flowpipe decoder options
func WithCredentials(credentials map[string]credential.Credential) parse.DecoderOption {
	return func(d parse.Decoder) {
		decoder, ok := d.(*FlowpipeModDecoder)
		if ok {
			decoder.Credentials = credentials
		}
	}
}

type FlowpipeModDecoder struct {
	parse.DecoderImpl
	Credentials map[string]credential.Credential
}

func NewFlowpipeModDecoder(opts ...parse.DecoderOption) parse.Decoder {
	d := &FlowpipeModDecoder{
		DecoderImpl: parse.NewDecoderImpl(),
	}
	d.DecodeFuncs[schema.BlockTypePipeline] = d.decodePipeline
	d.DecodeFuncs[schema.BlockTypeTrigger] = d.decodeTrigger
	// apply options
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *FlowpipeModDecoder) decodeStep(mod *modconfig.Mod, block *hcl.Block, parseCtx *parse.ModParseContext, pipelineHcl *resources.Pipeline) (resources.PipelineStep, hcl.Diagnostics) {

	stepType := block.Labels[0]
	stepName := block.Labels[1]

	step := resources.NewPipelineStep(stepType, stepName, pipelineHcl)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
			Subject:  &block.DefRange,
		}}
	}
	step.SetPipelineName(pipelineHcl.FullName)
	step.SetRange(block.DefRange.Ptr())

	pipelineStepBlockSchema := GetPipelineStepBlockSchema(stepType)
	if pipelineStepBlockSchema == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Pipeline step block schema not found for step " + stepType,
			Subject:  &block.DefRange,
		}}
	}

	stepOptions, diags := block.Body.Content(pipelineStepBlockSchema)

	if diags.HasErrors() {
		return step, diags
	}

	moreDiags := step.SetAttributes(stepOptions.Attributes, parseCtx.EvalCtx)
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	moreDiags = step.SetBlockConfig(stepOptions.Blocks, parseCtx.EvalCtx)
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	stepOutput := map[string]*resources.PipelineOutput{}

	outputBlocks := stepOptions.Blocks.ByType()[schema.BlockTypePipelineOutput]
	for _, outputBlock := range outputBlocks {
		attributes, moreDiags := outputBlock.Body.JustAttributes()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}

		if attr, exists := attributes[schema.AttributeTypeValue]; exists {

			o := &resources.PipelineOutput{
				Name: outputBlock.Labels[0],
			}

			expr := attr.Expr
			if len(expr.Variables()) > 0 {
				traversals := expr.Variables()
				for _, traversal := range traversals {
					parts := hclhelpers.TraversalAsStringSlice(traversal)
					if len(parts) > 0 {
						if parts[0] == schema.BlockTypePipelineStep {
							dependsOn := parts[1] + "." + parts[2]
							step.AppendDependsOn(dependsOn)
						} else if parts[0] == schema.BlockTypeCredential {

							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								step.AppendCredentialDependsOn(dependsOn)
							} else {
								dependsOn := parts[1] + "." + parts[2]
								step.AppendCredentialDependsOn(dependsOn)
							}
						} else if parts[0] == schema.BlockTypeConnection {
							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								step.AppendConnectionDependsOn(dependsOn)
							} else {
								dependsOn := parts[1] + "." + parts[2]
								step.AppendConnectionDependsOn(dependsOn)
							}
						} else {
							dependsOn := parts[0]
							step.AppendConnectionDependsOn(dependsOn)
						}
					}
				}
				o.UnresolvedValue = attr.Expr
			} else {
				ctyVal, _ := attr.Expr.Value(nil)
				val, _ := hclhelpers.CtyToGo(ctyVal)
				o.Value = val
			}

			stepOutput[o.Name] = o
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Missing value attribute",
				Subject:  &block.DefRange,
			})
		}

	}
	step.SetOutputConfig(stepOutput)

	if len(diags) == 0 {
		moreDiags := step.Validate()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
		}
	}

	return step, diags
}

func (d *FlowpipeModDecoder) decodePipelineParam(block *hcl.Block, parseCtx *parse.ModParseContext) (*resources.PipelineParam, hcl.Diagnostics) {
	o := &resources.PipelineParam{
		Name: block.Labels[0],
	}

	utils.LogTime(fmt.Sprintf("decode pipeline param %s start", o.Name))

	// because we want to use late binding for temp creds *and* the ability for pipeline param to define custom type,
	// we do the validation with with a list of temporary connections

	utils.LogTime(fmt.Sprintf("decode pipeline param %s start: set include late binding resources(true)", o.Name))
	parseCtx.SetIncludeLateBindingResources(true)
	utils.LogTime(fmt.Sprintf("decode pipeline param %s end: set include late binding resources(true)", o.Name))

	// be sure to revert the eval context to remove the temporary connections again
	defer func() {
		utils.LogTime(fmt.Sprintf("decode pipeline param %s start: set include late binding resources(false)", o.Name))
		parseCtx.SetIncludeLateBindingResources(false)
		utils.LogTime(fmt.Sprintf("decode pipeline param %s end: set include late binding resources(false)", o.Name))
	}()

	paramOptions, diags := block.Body.Content(resources.PipelineParamBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeType]; exists {
		ty, diags := parse.DecodeTypeExpression(attr)
		if diags.HasErrors() {
			return o, diags
		}

		o.Type = ty
		// get source data from eval context
		src := parseCtx.FileData[attr.Expr.Range().Filename]

		o.TypeString = parse.ExtractExpressionString(attr.Expr, src)
	} else {
		o.Type = cty.DynamicPseudoType
		o.TypeString = "any"
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeOptional]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Optional)
		diags = append(diags, valDiags...)
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeDefault]; exists {
		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		diags = append(diags, moreDiags...)
		if diags.HasErrors() {
			return o, diags
		}

		// Does the default value matches the specified type?
		utils.LogTime(fmt.Sprintf("decode pipeline param %s start: validate value matches type", o.Name))
		moreDiags = modconfig.ValidateValueMatchesType(ctyVal, o.Type, attr.Range.Ptr())
		utils.LogTime(fmt.Sprintf("decode pipeline param %s end: validate value matches type", o.Name))
		diags = append(diags, moreDiags...)
		if diags.HasErrors() {
			return o, diags
		}
		o.Default = ctyVal

	} else if o.Optional {
		o.Default = cty.NullVal(o.Type)
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeDescription]; exists {
		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			return o, diags
		}

		o.Description = ctyVal.AsString()
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeEnum]; exists {
		allowedTypes := []cty.Type{
			cty.String, cty.Bool, cty.Number,
			cty.List(cty.String), cty.List(cty.Bool), cty.List(cty.Number),
		}

		if !containsType(allowedTypes, o.Type) {
			return o, append(diags, createErrorDiagnostic("enum is only supported for string, bool, number, list of string, list of bool, list of number types", &attr.Range))
		}

		ctyVal, moreDiags := attr.Expr.Value(parseCtx.EvalCtx)
		if len(moreDiags) > 0 {
			return o, append(diags, moreDiags...)
		}

		if !hclhelpers.IsCollectionOrTuple(ctyVal.Type()) {
			return o, append(diags, createErrorDiagnostic("enum values must be a list", &attr.Range))
		}

		if !hclhelpers.IsEnumValueCompatibleWithType(o.Type, ctyVal) {
			return o, append(diags, createErrorDiagnostic("enum values type mismatched", &attr.Range))
		}

		if o.Default != cty.NilVal {
			if !hclhelpers.IsEnumValueCompatibleWithType(o.Default.Type(), ctyVal) {
				return o, append(diags, createErrorDiagnostic("param default value type mismatched with enum in pipeline param", &attr.Range))
			}
			if valid, err := hclhelpers.ValidateSettingWithEnum(o.Default, ctyVal); err != nil || !valid {
				return o, append(diags, createErrorDiagnostic("default value not in enum or error validating", &attr.Range))
			}
		}

		o.Enum = ctyVal

		enumGo, err := hclhelpers.CtyToGo(o.Enum)
		if err != nil {
			return o, append(diags, createErrorDiagnostic("error converting enum to go", &attr.Range))
		}

		enumGoSlice, ok := enumGo.([]any)
		if !ok {
			return o, append(diags, createErrorDiagnostic("enum is not a slice", &attr.Range))
		}

		o.EnumGo = enumGoSlice
	}

	if _, exists := paramOptions.Attributes[schema.AttributeTypeTags]; exists {
		valDiags := parse.DecodeProperty(paramOptions, "tags", &o.Tags, parseCtx.EvalCtx)
		diags = append(diags, valDiags...)
	}

	if attr, exists := paramOptions.Attributes[schema.AttributeTypeFormat]; exists {
		formatVal, moreDiags := parse.DecodeVarFormat(o.Type, attr, parseCtx)
		diags = append(diags, moreDiags...)
		if diags.HasErrors() {
			return o, diags
		}
		o.Format = formatVal
	} else if o.Type == cty.String {
		// if this is a string param, default to text
		o.Format = constants.VariableFormatText
	}

	utils.LogTime(fmt.Sprintf("decode pipeline param %s end", o.Name))

	return o, diags
}

func (d *FlowpipeModDecoder) decodeOutput(block *hcl.Block, parseCtx *parse.ModParseContext) (*resources.PipelineOutput, hcl.Diagnostics) {

	o := &resources.PipelineOutput{
		Name:  block.Labels[0],
		Range: block.DefRange.Ptr(),
	}

	outputOptions, diags := block.Body.Content(resources.PipelineOutputBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeDescription]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Description)
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
						if len(parts) >= 3 {
							dependsOn := parts[1] + "." + parts[2]
							o.AppendDependsOn(dependsOn)
						}
					} else if parts[0] == schema.BlockTypeCredential {

						if len(parts) == 2 {
							// dynamic references:
							// step "transform" "aws" {
							// 	value   = credential.aws[param.cred].env
							// }
							dependsOn := parts[1] + ".<dynamic>"
							o.AppendCredentialDependsOn(dependsOn)
						} else {
							dependsOn := parts[1] + "." + parts[2]
							o.AppendCredentialDependsOn(dependsOn)
						}
					} else if parts[0] == schema.BlockTypeConnection {
						dependsOn := parts[1] + "." + parts[2]
						o.AppendConnectionDependsOn(dependsOn)
					} else {
						dependsOn := parts[0]
						o.AppendConnectionDependsOn(dependsOn)
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

func (d *FlowpipeModDecoder) decodeTrigger(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {

	res := parse.NewDecodeResult()

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

	triggerHcl := resources.NewTrigger(block, parseCtx.CurrentMod, triggerType, triggerName)

	triggerSchema := GetTriggerBlockSchema(triggerType)
	if triggerSchema == nil {
		res.HandleDecodeDiags(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "invalid trigger type: " + triggerType,
				Subject:  &block.DefRange,
			},
		})
		return triggerHcl, res
	}

	triggerOptions, diags := block.Body.Content(triggerSchema)

	if diags.HasErrors() {
		res.HandleDecodeDiags(diags)
		return nil, res
	}

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

	diags = triggerHcl.Config.SetAttributes(parseCtx.CurrentMod, triggerHcl, triggerOptions.Attributes, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return triggerHcl, res
	}

	diags = triggerHcl.Config.SetBlocks(parseCtx.CurrentMod, triggerHcl, triggerOptions.Blocks, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return triggerHcl, res
	}

	var triggerParams []resources.PipelineParam
	for _, block := range triggerOptions.Blocks {
		if block.Type == schema.BlockTypeParam {
			param, diags := d.decodePipelineParam(block, parseCtx)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return triggerHcl, res
			}
			triggerParams = append(triggerParams, *param)
		}
	}

	body, ok := block.Body.(*hclsyntax.Body)
	if ok {
		triggerHcl.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
	} else {
		// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
		// but not the end line
		triggerHcl.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
	}

	// TODO K check this is ok
	//moreDiags := parseCtx.AddTrigger(triggerHcl)
	//res.AddDiags(moreDiags)

	triggerHcl.Params = triggerParams

	return triggerHcl, res
}

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func (d *FlowpipeModDecoder) decodePipeline(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {

	res := parse.NewDecodeResult()

	mod := parseCtx.CurrentMod
	// get shell pipelineHcl
	pipelineHcl := resources.NewPipeline(mod, block)

	utils.LogTime(fmt.Sprintf("decode pipeline %s start", pipelineHcl.FullName))

	pipelineOptions, diags := block.Body.Content(resources.PipelineBlockSchema)
	if diags.HasErrors() {
		res.HandleDecodeDiags(diags)
		return pipelineHcl, res
	}

	diags = pipelineHcl.SetAttributes(pipelineOptions.Attributes, parseCtx.EvalCtx)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
		return pipelineHcl, res
	}

	// use a map keyed by a string for fast lookup
	// we use an empty struct as the value type, so that
	// we don't use up unnecessary memory
	// foundOptions := map[string]struct{}{}
	for _, block := range pipelineOptions.Blocks {
		utils.LogTime(fmt.Sprintf("decode pipeline.block %s - %v start", block.Type, block.Labels))

		switch block.Type {
		case schema.BlockTypePipelineStep:
			step, diags := d.decodeStep(mod, block, parseCtx, pipelineHcl)
			if diags.HasErrors() {
				res.HandleDecodeDiags(diags)

				// Must also return the pipelineHcl even if it failed parsing, because later on the handling of "unresolved blocks" expect
				// the resource to be there
				return pipelineHcl, res
			}

			body, ok := block.Body.(*hclsyntax.Body)
			if ok {
				step.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
			} else {
				// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
				// but not the end line
				step.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
			}

			pipelineHcl.Steps = append(pipelineHcl.Steps, step)

		case schema.BlockTypePipelineOutput:
			output, cfgDiags := d.decodeOutput(block, parseCtx)
			diags = append(diags, cfgDiags...)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return pipelineHcl, res
			}

			if output != nil {

				// check for duplicate output names
				if len(pipelineHcl.OutputConfig) > 0 {
					for _, o := range pipelineHcl.OutputConfig {
						if o.Name == output.Name {
							diags = append(diags, &hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  fmt.Sprintf("duplicate output name '%s' - output names must be unique", output.Name),
								Subject:  &block.DefRange,
							})
							res.HandleDecodeDiags(diags)
							return pipelineHcl, res
						}
					}
				}

				pipelineHcl.OutputConfig = append(pipelineHcl.OutputConfig, *output)
			}

		case schema.BlockTypeParam:
			pipelineParam, moreDiags := d.decodePipelineParam(block, parseCtx)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				res.HandleDecodeDiags(diags)
				return pipelineHcl, res
			}

			if pipelineParam != nil {

				// check for duplicate pipeline parameters names
				if len(pipelineHcl.Params) > 0 {
					p := pipelineHcl.GetParam(pipelineParam.Name)

					if p != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  fmt.Sprintf("duplicate pipeline parameter name '%s' - parameter names must be unique", pipelineParam.Name),
							Subject:  &block.DefRange,
						})
						res.HandleDecodeDiags(diags)
						return pipelineHcl, res
					}
				}

				pipelineHcl.Params = append(pipelineHcl.Params, *pipelineParam)
			}

		default:
			// this should never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for workspace profiles", block.Type),
				Subject:  &block.DefRange,
			})
		}

		utils.LogTime(fmt.Sprintf("decode pipeline.block %s - %v end", block.Type, block.Labels))
	}

	diags = validatePipelineSteps(pipelineHcl)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)

		return pipelineHcl, res
	}

	handlePipelineDecodeResult(pipelineHcl, res, block, parseCtx)
	diags = validatePipelineDependencies(pipelineHcl, d.Credentials, parseCtx.PipelingConnections)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)

		// Must also return the pipelineHcl even if it failed parsing, because later on the handling of "unresolved blocks" expect
		// the resource to be there
		return pipelineHcl, res
	}

	body, ok := block.Body.(*hclsyntax.Body)
	if ok {
		pipelineHcl.SetFileReference(block.DefRange.Filename, body.SrcRange.Start.Line, body.EndRange.Start.Line)
	} else {
		// This shouldn't happen, but if it does, try our best effort to set the file reference. It will get the start line correctly
		// but not the end line
		pipelineHcl.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)
	}

	utils.LogTime(fmt.Sprintf("decode pipeline %s end", pipelineHcl.FullName))
	return pipelineHcl, res
}

// Checks if the given type is in the allowed list
func containsType(allowedTypes []cty.Type, typ cty.Type) bool {
	for _, t := range allowedTypes {
		if t == typ {
			return true
		}
	}
	return false
}

// Creates an HCL error diagnostic
func createErrorDiagnostic(summary string, subject *hcl.Range) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  summary,
		Subject:  subject,
	}
}

func validatePipelineSteps(pipelineHcl *resources.Pipeline) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	stepMap := map[string]bool{}

	for _, step := range pipelineHcl.Steps {

		if _, ok := stepMap[step.GetFullyQualifiedName()]; ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("duplicate step name '%s' - step names must be unique", step.GetFullyQualifiedName()),
				Subject:  step.GetRange(),
			})
			continue
		}

		stepMap[step.GetFullyQualifiedName()] = true

		moreDiags := step.Validate()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}
	}

	return diags
}

func validatePipelineDependencies(pipelineHcl *resources.Pipeline, credentials map[string]credential.Credential, connections map[string]connection.PipelingConnection) hcl.Diagnostics {
	var diags hcl.Diagnostics

	var stepRegisters []string
	for _, step := range pipelineHcl.Steps {
		stepRegisters = append(stepRegisters, step.GetFullyQualifiedName())
	}

	var credentialRegisters []string
	availableCredentialTypes := map[string]bool{}
	for k := range credentials {
		parts := strings.Split(k, ".")
		if len(parts) != 2 {
			continue
		}

		// Add the credential to the register
		credentialRegisters = append(credentialRegisters, k)

		// List out the supported credential types
		availableCredentialTypes[parts[0]] = true
	}

	var credentialTypes []string
	for k := range availableCredentialTypes {
		credentialTypes = append(credentialTypes, k)
	}

	var connectionRegisters []string
	availableConnectionTypes := map[string]bool{}
	for k := range connections {
		parts := strings.Split(k, ".")
		if len(parts) != 2 {
			continue
		}

		// Add the connection to the register
		connectionRegisters = append(connectionRegisters, k)

		// List out the supported connection types
		availableConnectionTypes[parts[0]] = true
	}

	var connectionTypes []string
	for k := range availableConnectionTypes {
		connectionTypes = append(connectionTypes, k)
	}

	for _, step := range pipelineHcl.Steps {
		dependsOn := step.GetDependsOn()

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', step '%s' does not exist in pipeline %s", dep, dep, pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid steps are: %s", strings.Join(stepRegisters, ", ")),
					Subject:  step.GetRange(),
				})
			}
		}

		credentialDependsOn := step.GetCredentialDependsOn()
		for _, dep := range credentialDependsOn {
			// Check if the credential type is supported, if <dynamic>
			parts := strings.Split(dep, ".")
			if len(parts) != 2 {
				continue
			}

			if parts[1] == "<dynamic>" {
				if !availableCredentialTypes[parts[0]] {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("invalid depends_on '%s', credential type '%s' not supported in pipeline %s", dep, parts[0], pipelineHcl.Name()),
						Detail:   fmt.Sprintf("valid credential types are: %s", strings.Join(credentialTypes, ", ")),
						Subject:  step.GetRange(),
					})
				}
				continue
			}

			if !helpers.StringSliceContains(credentialRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', credential does not exist in pipeline %s", dep, pipelineHcl.Name()),
					Detail:   fmt.Sprintf("valid credentials are: %s", strings.Join(credentialRegisters, ", ")),
					Subject:  step.GetRange(),
				})
			}
		}

		connectionDependsOn := step.GetConnectionDependsOn()
		for _, dep := range connectionDependsOn {
			// Check if the credential type is supported, if <dynamic>
			parts := strings.Split(dep, ".")
			if len(parts) != 2 {
				continue
			}

			if parts[1] == "<dynamic>" {
				if !availableConnectionTypes[parts[0]] {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("invalid depends_on '%s', connection type '%s' not supported in pipeline %s", dep, parts[0], pipelineHcl.Name()),
						Detail:   fmt.Sprintf("valid connection types are: %s", strings.Join(connectionTypes, ", ")),
						Subject:  step.GetRange(),
					})
				}
				continue
			}

			if !helpers.StringSliceContains(connectionRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s', connection does not exist in pipeline %s", dep, pipelineHcl.Name()),
					Subject:  step.GetRange(),
				})
			}
		}

	}

	for _, outputConfig := range pipelineHcl.OutputConfig {
		dependsOn := outputConfig.DependsOn

		for _, dep := range dependsOn {
			if !helpers.StringSliceContains(stepRegisters, dep) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid depends_on '%s' in output block, '%s' does not exist in pipeline %s", dep, dep, pipelineHcl.Name()),
					Subject:  outputConfig.Range,
				})
			}
		}
	}

	return diags
}

func handlePipelineDecodeResult(resource *resources.Pipeline, res *parse.DecodeResult, block *hcl.Block, parseCtx *parse.ModParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded(block, parseCtx)
		res.AddDiags(moreDiags)

		// TODO K verify no longer needed
		//moreDiags = parseCtx.AddPipeline(resource)
		//res.AddDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.AddDiags(moreDiags)
	}
}

func GetPipelineStepBlockSchema(stepType string) *hcl.BodySchema {
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return resources.PipelineStepHttpBlockSchema
	case schema.BlockTypePipelineStepSleep:
		return resources.PipelineStepSleepBlockSchema
	case schema.BlockTypePipelineStepEmail:
		return resources.PipelineStepEmailBlockSchema
	case schema.BlockTypePipelineStepTransform:
		return resources.PipelineStepTransformBlockSchema
	case schema.BlockTypePipelineStepQuery:
		return resources.PipelineStepQueryBlockSchema
	case schema.BlockTypePipelineStepPipeline:
		return resources.PipelineStepPipelineBlockSchema
	case schema.BlockTypePipelineStepFunction:
		return resources.PipelineStepFunctionBlockSchema
	case schema.BlockTypePipelineStepContainer:
		return resources.PipelineStepContainerBlockSchema
	case schema.BlockTypePipelineStepInput:
		return resources.PipelineStepInputBlockSchema
	case schema.BlockTypePipelineStepMessage:
		return resources.PipelineStepMessageBlockSchema
	default:
		return nil
	}
}

func GetTriggerBlockSchema(triggerType string) *hcl.BodySchema {
	switch triggerType {
	case schema.TriggerTypeSchedule:
		return resources.TriggerScheduleBlockSchema
	case schema.TriggerTypeQuery:
		return resources.TriggerQueryBlockSchema
	case schema.TriggerTypeHttp:
		return resources.TriggerHttpBlockSchema
	default:
		return nil
	}
}

func GetIntegrationBlockSchema(integrationType string) *hcl.BodySchema {
	switch integrationType {
	case schema.IntegrationTypeSlack:
		return resources.IntegrationSlackBlockSchema
	case schema.IntegrationTypeEmail:
		return resources.IntegrationEmailBlockSchema
	case schema.IntegrationTypeMsTeams:
		return resources.IntegrationTeamsBlockSchema
	default:
		return nil
	}
}
