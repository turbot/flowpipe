package fpconfig

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser"
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
		return fperr.InternalWithMessage(fmt.Sprintf("%v", val))
	}
}

func LoadFlowpipeConfig(ctx context.Context, configPath string) (map[string]*types.Pipeline, error) {
	// create profile map to populate
	pipelineMap := map[string]*types.Pipeline{}

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
		return pipelineMap, nil
	}

	fileData, diags := pipeparser.LoadFileData(flowpipeConfigFilePaths...)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	if len(fileData) != len(flowpipeConfigFilePaths) {
		return nil, fperr.InternalWithMessage("Failed to load all pipeline files")
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

	// content.Blocks is the list of all pipeline blocks found in the pipeline files
	parseCtx := NewFlowpipeConfigParseContext(ctx, configPath)
	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	pipelines, err := parseAllFlowipeConfig(parseCtx)
	if err != nil {
		return nil, fperr.Internal(err)
	}

	return pipelines, nil
}

func LoadPipelines(ctx context.Context, configPath string) (map[string]*types.Pipeline, error) {
	pipelineMap, err := LoadFlowpipeConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}
	return pipelineMap, nil
}

func parseAllFlowipeConfig(parseCtx *FlowpipeConfigParseContext) (map[string]*types.Pipeline, error) {
	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags := decodeFlowpipeConfigBlocks(parseCtx)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("Failed to decode pipelines", diags)
		}

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			// log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			return nil, fmt.Errorf("failed to resolve workspace profile dependencies after %d attempts\nDependencies:\n%s", attempts+1, str)
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	return parseCtx.PipelineHcls, nil

}

func decodeFlowpipeConfigBlocks(parseCtx *FlowpipeConfigParseContext) hcl.Diagnostics {
	pipelines := map[string]*types.Pipeline{}
	// triggers := map[string]*types.Trigger{}

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

	// blocksToDecode contains the number of pipeline block we found in all the files
	// from the given input directory.
	//
	// each "block" is the pipeline HCL block that we need to decode into a Go Struct
	for _, block := range blocksToDecode {
		switch block.Type {
		case schema.BlockTypePipeline:
			pipelineHcl, res := decodePipeline(block, parseCtx)

			if res.Success() {
				pipelines[pipelineHcl.Name] = pipelineHcl
			}
			diags = append(diags, res.Diags...)
			// case schema.BlockTypeTrigger:
			// 	triggerHcl, res := decodeTrigger(block, parseCtx)
			// 	if res.Success() {
			// 		triggers[triggerHcl.Name] = triggerHcl
			// 	}
			// 	diags = append(diags, res.Diags...)
		}
	}
	return diags
}

// func decodeTrigger(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*types.Trigger, *pipeparser.DecodeResult) {
// 	return nil, nil
// }

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*types.Pipeline, *pipeparser.DecodeResult) {
	res := pipeparser.NewDecodeResult()
	// get shell pipelineHcl
	pipelineHcl := types.NewPipelineHcl(block)

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

func validatePipelineDependencies(pipelineHcl *types.Pipeline) hcl.Diagnostics {
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

func handlePipelineDecodeResult(resource *types.Pipeline, res *pipeparser.DecodeResult, block *hcl.Block, parseCtx *FlowpipeConfigParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded()
		res.AddDiags(moreDiags)

		moreDiags = parseCtx.AddResource(resource)
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

type FlowpipeConfigParseContext struct {
	pipeparser.ParseContext
	PipelineHcls map[string]*types.Pipeline
	TriggerHcls  map[string]*types.Trigger
}

func (c *FlowpipeConfigParseContext) buildEvalContext() {
	vars := map[string]cty.Value{}
	c.ParseContext.BuildEvalContext(vars)
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (c *FlowpipeConfigParseContext) AddResource(pipelineHcl *types.Pipeline) hcl.Diagnostics {
	// ctyVal, err := pipelineHcl.AsCtyValue()
	// if err != nil {
	// 	return hcl.Diagnostics{&hcl.Diagnostic{
	// 		Severity: hcl.DiagError,
	// 		Summary:  fmt.Sprintf("failed to convert Pipeline '%s' to its cty value", pipelineHcl.Name),
	// 		Detail:   err.Error(),
	// 		// TODO: fix this
	// 		// Subject:  &workspaceProfile.DeclRange,
	// 		// Subject: "change me",
	// 	}}
	// }

	c.PipelineHcls[pipelineHcl.Name] = pipelineHcl
	// c.valueMap[pipelineHcl.Name] = ctyVal

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, pipelineHcl.Name)

	c.buildEvalContext()

	return nil
}

func NewFlowpipeConfigParseContext(ctx context.Context, rootEvalPath string) *FlowpipeConfigParseContext {
	parseContext := pipeparser.NewParseContext(ctx, rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done

	c := &FlowpipeConfigParseContext{
		ParseContext: parseContext,
		PipelineHcls: make(map[string]*types.Pipeline),
		TriggerHcls:  make(map[string]*types.Trigger),
	}

	c.buildEvalContext()

	return c
}
