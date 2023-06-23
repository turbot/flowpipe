package pipeline_hcl

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/options"
	filehelpers "github.com/turbot/go-kit/files"
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

func LoadPipelines(ctx context.Context, pipelinePath string) (pipelineMap map[string]*types.PipelineHcl, err error) {

	// create profile map to populate
	pipelineMap = map[string]*types.PipelineHcl{}

	configPaths, err := filehelpers.ListFiles(pipelinePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{pipeparser.PipelineExtension}),
	})
	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return pipelineMap, nil
	}

	fileData, diags := pipeparser.LoadFileData(configPaths...)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	body, diags := pipeparser.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := body.Content(PipelineBlockSchema)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	parseCtx := NewPipelineParseContext(pipelinePath)
	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	pipelines, err := parsePipelines(parseCtx)
	if err != nil {
		return nil, fperr.Internal(err)
	}

	return pipelines, nil
}

func parsePipelines(parseCtx *PipelineParseContext) (map[string]*types.PipelineHcl, error) {
	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		_, diags := decodePipelineHcls(parseCtx)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("Failed to decode all workspace profile files", diags)
		}

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
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

func decodePipelineHcls(parseCtx *PipelineParseContext) (map[string]*types.PipelineHcl, hcl.Diagnostics) {
	profileMap := map[string]*types.PipelineHcl{}

	var diags hcl.Diagnostics
	blocksToDecode, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return nil, diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocksToDecode {
		if block.Type == modconfig.BlockTypePipeline {
			pipelineHcl, res := decodePipeline(block, parseCtx)

			if res.Success() {
				// success - add to map
				profileMap[pipelineHcl.Name] = pipelineHcl
			}
			diags = append(diags, res.Diags...)
		}
	}
	return profileMap, diags
}

func decodePipeline(block *hcl.Block, parseCtx *PipelineParseContext) (*types.PipelineHcl, *pipeparser.DecodeResult) {
	res := pipeparser.NewDecodeResult()
	// get shell resource
	resource := types.NewPipelineHcl(block)

	// do a partial decode to get options blocks into pipelineOptions, with all other attributes in rest
	pipelineOptions, rest, diags := block.Body.PartialContent(PipelineBlockSchema)
	if diags.HasErrors() {
		res.HandleDecodeDiags(diags)
		return nil, res
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, resource)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
	}
	// use a map keyed by a string for fast lookup
	// we use an empty struct as the value type, so that
	// we don't use up unnecessary memory
	foundOptions := map[string]struct{}{}
	for _, block := range pipelineOptions.Blocks {
		switch block.Type {
		case "options":
			optionsBlockType := block.Labels[0]
			if _, found := foundOptions[optionsBlockType]; found {
				// fail
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Subject:  &block.DefRange,
					Summary:  fmt.Sprintf("Duplicate options type '%s'", optionsBlockType),
				})
			}
			opts, moreDiags := decodePipelineHclOption(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			moreDiags = resource.SetOptions(opts, block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
			foundOptions[optionsBlockType] = struct{}{}
		default:
			// this should never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for workspace profiles", block.Type),
				Subject:  &block.DefRange,
			})
		}
	}

	handlePipelineDecodeResult(resource, res, block, parseCtx)
	return resource, res
}

func handlePipelineDecodeResult(resource *types.PipelineHcl, res *pipeparser.DecodeResult, block *hcl.Block, parseCtx *PipelineParseContext) {
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

// decodeWorkspaceProfileOption decodes an options block as a workspace profile property
// setting the necessary overrides for special handling of the "dashboard" option which is different
// from the global "dashboard" option
func decodePipelineHclOption(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	// return pipeparser.DecodeOptions(block, pipeparser.WithOverride(constants.CmdNameDashboard, &options.WorkspaceProfileDashboard{}))
	return pipeparser.DecodeOptions(block)
}

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},

	// TODO: what's this?
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "pipeline",
			LabelNames: []string{"name"},
		},
	},
}

type PipelineParseContext struct {
	pipeparser.ParseContext
	PipelineHcls map[string]*types.PipelineHcl
	valueMap     map[string]cty.Value
}

func (c *PipelineParseContext) buildEvalContext() {
	// rebuild the eval context
	// build a map with a single key - workspace
	vars := map[string]cty.Value{
		"pipeline": cty.ObjectVal(c.valueMap),
	}
	c.ParseContext.BuildEvalContext(vars)

}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (c *PipelineParseContext) AddResource(workspaceProfile *types.PipelineHcl) hcl.Diagnostics {
	ctyVal, err := workspaceProfile.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert workspaceProfile '%s' to its cty value", workspaceProfile.Name),
			Detail:   err.Error(),
			// TODO: fix this
			// Subject:  &workspaceProfile.DeclRange,
			// Subject: "change me",
		}}
	}

	c.PipelineHcls[workspaceProfile.Name] = workspaceProfile
	c.valueMap[workspaceProfile.Name] = ctyVal

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, workspaceProfile.Name)

	c.buildEvalContext()

	return nil
}

func NewPipelineParseContext(rootEvalPath string) *PipelineParseContext {
	parseContext := pipeparser.NewParseContext(rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done
	//parseContext.BlockTypes = []string{modconfig.BlockTypeWorkspaceProfile}
	c := &PipelineParseContext{
		ParseContext: parseContext,
		PipelineHcls: make(map[string]*types.PipelineHcl),
		valueMap:     make(map[string]cty.Value),
	}

	c.buildEvalContext()

	return c
}
