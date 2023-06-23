package pipeline_hcl

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/options"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/zclconf/go-cty/cty"
)

/*
	type WorkspaceProfile struct {
		ProfileName       string            `hcl:"name,label" cty:"name"`
*/
type PipelineHcl struct {
	Name    string            `hcl:"name,label" cty:"name"`
	Output  *string           `hcl:"output" cty:"output"`
	Steps   []PipelineHclStep `hcl:"step,block" cty:"step"`
	RawBody hcl.Body          `hcl:",remain"`
}

type PipelineHclStep struct {
	Type string `hcl:"type,label" cty:"type"`
	Name string `hcl:"name,label" cty:"name"`

	// Unparsed HCL for the step configuration. Each step type has differing structure.
	Config hcl.Body `hcl:",remain"`
}

func (p *PipelineHcl) CtyValue() (cty.Value, error) {
	return pipeparser.GetCtyValue(p)
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *PipelineHcl) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {

	var diags hcl.Diagnostics
	switch o := opts.(type) {
	// case *options.Query:
	// 	if p.QueryOptions != nil {
	// 		diags = append(diags, duplicateOptionsBlockDiag(block))
	// 	}
	// 	p.QueryOptions = o
	// case *options.Check:
	// 	if p.CheckOptions != nil {
	// 		diags = append(diags, duplicateOptionsBlockDiag(block))
	// 	}
	// 	p.CheckOptions = o
	// case *options.WorkspaceProfileDashboard:
	// 	if p.DashboardOptions != nil {
	// 		diags = append(diags, duplicateOptionsBlockDiag(block))
	// 	}
	// 	p.DashboardOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (p *PipelineHcl) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *PipelineHcl) setBaseProperties() {
	// 	if p.Base == nil {
	// 		return
	// 	}

	// 	if p.CloudHost == nil {
	// 		p.CloudHost = p.Base.CloudHost
	// 	}
	// 	if p.CloudToken == nil {
	// 		p.CloudToken = p.Base.CloudToken
	// 	}
	// 	if p.InstallDir == nil {
	// 		p.InstallDir = p.Base.InstallDir
	// 	}
	// 	if p.ModLocation == nil {
	// 		p.ModLocation = p.Base.ModLocation
	// 	}
	// 	if p.SnapshotLocation == nil {
	// 		p.SnapshotLocation = p.Base.SnapshotLocation
	// 	}
	// 	if p.WorkspaceDatabase == nil {
	// 		p.WorkspaceDatabase = p.Base.WorkspaceDatabase
	// 	}
	// 	if p.QueryTimeout == nil {
	// 		p.QueryTimeout = p.Base.QueryTimeout
	// 	}
	// 	if p.SearchPath == nil {
	// 		p.SearchPath = p.Base.SearchPath
	// 	}
	// 	if p.SearchPathPrefix == nil {
	// 		p.SearchPathPrefix = p.Base.SearchPathPrefix
	// 	}
	// 	if p.Watch == nil {
	// 		p.Watch = p.Base.Watch
	// 	}
	// 	if p.MaxParallel == nil {
	// 		p.MaxParallel = p.Base.MaxParallel
	// 	}
	// 	if p.Introspection == nil {
	// 		p.Introspection = p.Base.Introspection
	// 	}
	// 	if p.Input == nil {
	// 		p.Input = p.Base.Input
	// 	}
	// 	if p.Progress == nil {
	// 		p.Progress = p.Base.Progress
	// 	}
	// 	if p.Theme == nil {
	// 		p.Theme = p.Base.Theme
	// 	}
	// 	if p.Cache == nil {
	// 		p.Cache = p.Base.Cache
	// 	}
	// 	if p.CacheTTL == nil {
	// 		p.CacheTTL = p.Base.CacheTTL
	// 	}

	// 	// nested inheritance strategy:
	// 	//
	// 	// if my nested struct is a nil
	// 	//		-> use the base struct
	// 	//
	// 	// if I am not nil (and base is not nil)
	// 	//		-> only inherit the properties which are nil in me and not in base
	// 	//
	// 	if p.QueryOptions == nil {
	// 		p.QueryOptions = p.Base.QueryOptions
	// 	} else {
	// 		p.QueryOptions.SetBaseProperties(p.Base.QueryOptions)
	// 	}
	// 	if p.CheckOptions == nil {
	// 		p.CheckOptions = p.Base.CheckOptions
	// 	} else {
	// 		p.CheckOptions.SetBaseProperties(p.Base.CheckOptions)
	// 	}
	// 	if p.DashboardOptions == nil {
	// 		p.DashboardOptions = p.Base.DashboardOptions
	// 	} else {
	// 		p.DashboardOptions.SetBaseProperties(p.Base.DashboardOptions)
	// 	}
}

func NewPipelineHcl(block *hcl.Block) *PipelineHcl {
	return &PipelineHcl{
		Name: block.Labels[0],
	}
}

// ToError formats the supplied value as an error (or just returns it if already an error)
func ToError(val interface{}) error {
	if e, ok := val.(error); ok {
		return e
	} else {
		return fperr.InternalWithMessage(fmt.Sprintf("%v", val))
	}
}

func LoadWorkspacePipelines(pipelinePath string) (pipelineMap map[string]*PipelineHcl, err error) {

	// create profile map to populate
	pipelineMap = map[string]*PipelineHcl{}

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
	return parsePipelines(parseCtx)

}

func parsePipelines(parseCtx *PipelineParseContext) (map[string]*PipelineHcl, error) {
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

	return parseCtx.pipelineHcls, nil

}

func decodePipelineHcls(parseCtx *PipelineParseContext) (map[string]*PipelineHcl, hcl.Diagnostics) {
	profileMap := map[string]*PipelineHcl{}

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

func decodePipeline(block *hcl.Block, parseCtx *PipelineParseContext) (*PipelineHcl, *pipeparser.DecodeResult) {
	res := pipeparser.NewDecodeResult()
	// get shell resource
	resource := NewPipelineHcl(block)

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

func handlePipelineDecodeResult(resource *PipelineHcl, res *pipeparser.DecodeResult, block *hcl.Block, parseCtx *PipelineParseContext) {
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
	pipelineHcls map[string]*PipelineHcl
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
func (c *PipelineParseContext) AddResource(workspaceProfile *PipelineHcl) hcl.Diagnostics {
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

	c.pipelineHcls[workspaceProfile.Name] = workspaceProfile
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
		pipelineHcls: make(map[string]*PipelineHcl),
		valueMap:     make(map[string]cty.Value),
	}

	c.buildEvalContext()

	return c
}
