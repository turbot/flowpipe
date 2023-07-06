package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/configschema"
	"github.com/turbot/flowpipe/pipeparser/constants"
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

	pipelineFilePaths, err := filehelpers.ListFiles(pipelinePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{pipeparser.PipelineExtension}),
	})
	if err != nil {
		return nil, err
	}

	// pipelineFilePaths is the list of all pipeline files found in the pipelinePath
	if len(pipelineFilePaths) == 0 {
		return pipelineMap, nil
	}

	fileData, diags := pipeparser.LoadFileData(pipelineFilePaths...)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	if len(fileData) != len(pipelineFilePaths) {
		return nil, fperr.InternalWithMessage("Failed to load all pipeline files")
	}

	// Each file in the pipelineFilePaths is parsed and the result is stored in the bodies variable
	// bodies.data length should be the same with pipelineFilePaths length
	bodies, diags := pipeparser.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := bodies.Content(PipelineBlockSchema)
	if diags.HasErrors() {
		return nil, pipeparser.DiagsToError("Failed to load workspace profiles", diags)
	}

	// content.Blocks is the list of all pipeline blocks found in the pipeline files
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

	// blocksToDecode contains the number of pipeline block we found in all the files
	// from the given input directory.
	//
	// each "block" is the pipeline HCL block that we need to decode into a Go Struct
	for _, block := range blocksToDecode {
		if block.Type == configschema.BlockTypePipeline {
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

// TODO: validation - if you specify invalid depends_on it doesn't error out
// TODO: validation - invalid name?
func decodePipeline(block *hcl.Block, parseCtx *PipelineParseContext) (*types.PipelineHcl, *pipeparser.DecodeResult) {
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
		case configschema.BlockTypePipelineStep:
			stepType := block.Labels[0]
			stepName := block.Labels[1]

			step := types.NewPipelineStep(stepType, stepName)
			if step == nil {
				res.HandleDecodeDiags(hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("Invalid pipeline step type %s", stepType),
					},
				})
				return nil, res
			}

			stepOptions, rest, diags := block.Body.PartialContent(GetPipelineStepBlockSchema(stepType))

			if diags.HasErrors() {
				res.HandleDecodeDiags(diags)
				return nil, res
			}

			diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, step)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return nil, res
			}

			diags = step.SetAttributes(stepOptions.Attributes)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return nil, res
			}

			pipelineHcl.Steps = append(pipelineHcl.Steps, step)

		case configschema.BlockTypePipelineOutput:
			override := false
			output, cfgDiags := decodeOutputBlock(block, override)
			diags = append(diags, cfgDiags...)
			if len(diags) > 0 {
				res.HandleDecodeDiags(diags)
				return nil, res
			}

			if output != nil {
				pipelineHcl.HclOutputs = append(pipelineHcl.HclOutputs, output)
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
	return pipelineHcl, res
}

func decodeOutputBlock(block *hcl.Block, override bool) (*types.Output, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	o := &types.Output{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	schema := PipelineOutputBlockSchema
	// if override {
	// 	schema = schemaForOverrides(schema)
	// }

	content, moreDiags := block.Body.Content(schema)
	diags = append(diags, moreDiags...)

	if !hclsyntax.ValidIdentifier(o.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid output name",
			Detail:   constants.BadIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Description)
		diags = append(diags, valDiags...)
		o.DescriptionSet = true
	}

	if attr, exists := content.Attributes["value"]; exists {
		o.Expr = attr.Expr
	}

	if attr, exists := content.Attributes["sensitive"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Sensitive)
		diags = append(diags, valDiags...)
		o.SensitiveSet = true
	}

	// TODO: depends_on for output?
	// if attr, exists := content.Attributes["depends_on"]; exists {
	// 	deps, depsDiags := decodeDependsOn(attr)
	// 	diags = append(diags, depsDiags...)
	// 	o.DependsOn = append(o.DependsOn, deps...)
	// }

	// TODO: do we need this? The code is lifted from Terraform
	// for _, block := range content.Blocks {
	// 	switch block.Type {
	// 	case "precondition":
	// 		cr, moreDiags := decodeCheckRuleBlock(block, override)
	// 		diags = append(diags, moreDiags...)
	// 		o.Preconditions = append(o.Preconditions, cr)
	// 	case "postcondition":
	// 		diags = append(diags, &hcl.Diagnostic{
	// 			Severity: hcl.DiagError,
	// 			Summary:  "Postconditions are not allowed",
	// 			Detail:   "Output values can only have preconditions, not postconditions.",
	// 			Subject:  block.TypeRange.Ptr(),
	// 		})
	// 	default:
	// 		// The cases above should be exhaustive for all block types
	// 		// defined in the block type schema, so this shouldn't happen.
	// 		panic(fmt.Sprintf("unexpected lifecycle sub-block type %q", block.Type))
	// 	}
	// }

	return o, diags
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

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeDescription,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       configschema.BlockTypePipeline,
			LabelNames: []string{configschema.LabelName},
		},
		{
			Type:       configschema.BlockTypePipelineStep,
			LabelNames: []string{configschema.LabelType, configschema.LabelName},
		},
		{
			Type:       configschema.BlockTypePipelineOutput,
			LabelNames: []string{configschema.LabelName},
		},
	},
}

var PipelineOutputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name:     "value",
			Required: true,
		},
		{
			Name: "depends_on",
		},
		{
			Name: "sensitive",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "precondition"},
		{Type: "postcondition"},
	},
}

var PipelineStepHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeUrl,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}

var PipelineStepSleepBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeDuration,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}
var PipelineStepEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeTo,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}

func GetPipelineStepBlockSchema(stepType string) *hcl.BodySchema {
	switch stepType {
	case configschema.BlockTypePipelineStepHttp:
		return PipelineStepHttpBlockSchema
	case configschema.BlockTypePipelineStepSleep:
		return PipelineStepSleepBlockSchema
	case configschema.BlockTypePipelineStepEmail:
		return PipelineStepEmailBlockSchema
	default:
		return nil
	}
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
