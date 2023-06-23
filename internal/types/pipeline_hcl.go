package types

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/options"
	"github.com/zclconf/go-cty/cty"
)

func NewPipelineHcl(block *hcl.Block) *PipelineHcl {
	return &PipelineHcl{
		Name: block.Labels[0],
	}
}

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

func (p *PipelineHcl) GetStep(stepName string) *PipelineHclStep {
	for i := 0; i < len(p.Steps); i++ {
		if p.Steps[i].Name == stepName {
			return &p.Steps[i]
		}
	}
	return nil
}

type PipelineHclStep struct {
	Type string `hcl:"type,label" cty:"type"`
	Name string `hcl:"name,label" cty:"name"`

	Input     string
	DependsOn []string
	For       string
	Error     PipelineStepError

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
