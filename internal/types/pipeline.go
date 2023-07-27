package types

import (
	"encoding/json"
	"fmt"
	"reflect"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/options"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/flowpipe/pipeparser/terraform/configs"
	"github.com/zclconf/go-cty/cty"
)

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []Pipeline `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID           string `json:"execution_id"`
	PipelineExecutionID   string `json:"pipeline_execution_id"`
	ParentStepExecutionID string `json:"parent_step_execution_id"`
}

type CmdPipeline struct {
	Command string `json:"command" binding:"required,oneof=run"`
}

func NewPipelineHcl(block *hcl.Block) *Pipeline {
	return &Pipeline{
		Name:   block.Labels[0],
		Params: map[string]*configs.Variable{},
	}
}

// Pipeline represents a "pipeline" block in an flowpipe HCL (*.fp) file
//
// Note that this Pipeline definition is different that the pipeline that is running. This definition
// contains unresolved expressions (mostly in steps), how to handle errors etc but not the actual Pipeline
// execution data.
type Pipeline struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty" hcl:"description,optional" cty:"description"`

	// Unparsed HCL body, needed so we can de-code the step HCL into the correct struct
	RawBody hcl.Body `json:"-" hcl:",remain"`

	// Unparsed JSON raw message, needed so we can unmarshall the step JSON into the correct struct
	StepsRawJson json.RawMessage `json:"-"`

	Steps []IPipelineStep `json:"steps"`

	Outputs []PipelineOutput `json:"outputs"`

	Params map[string]*configs.Variable
}

func (p *Pipeline) GetStep(stepFullyQualifiedName string) IPipelineStep {
	for i := 0; i < len(p.Steps); i++ {
		if p.Steps[i].GetFullyQualifiedName() == stepFullyQualifiedName {
			return p.Steps[i]
		}
	}
	return nil
}

func (p *Pipeline) CtyValue() (cty.Value, error) {
	return pipeparser.GetCtyValue(p)
}

// TODO: this only works with default value, we need to modify this
// TODO: so it takes external parameter
func (p *Pipeline) ParamsAsCty() (cty.Value, error) {

	params := map[string]cty.Value{}

	for k, v := range p.Params {
		params[k] = v.Default
	}

	return cty.ObjectVal(params), nil
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *Pipeline) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {

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

func (p *Pipeline) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *Pipeline) setBaseProperties() {

}

func (ph *Pipeline) UnmarshalJSON(data []byte) error {
	// Define an auxiliary type to decode the JSON and capture the value of the 'ISteps' field
	type Aux struct {
		Name        string          `json:"name"`
		Description *string         `json:"description,omitempty"`
		Output      *string         `json:"output,omitempty"`
		Raw         json.RawMessage `json:"-"`
		ISteps      json.RawMessage `json:"steps"`
	}

	aux := Aux{ISteps: json.RawMessage([]byte("null"))} // Provide a default value for 'ISteps' field
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Assign values to the fields of the main struct
	ph.Name = aux.Name
	ph.Description = aux.Description
	ph.StepsRawJson = []byte(aux.Raw)

	// Determine the concrete type of 'ISteps' based on the data present in the JSON
	if aux.ISteps != nil && string(aux.ISteps) != "null" {
		// Replace the JSON array of 'ISteps' with the desired concrete type
		var stepSlice []json.RawMessage
		if err := json.Unmarshal(aux.ISteps, &stepSlice); err != nil {
			return err
		}

		// Iterate over the stepSlice and determine the concrete type of each step
		for _, stepData := range stepSlice {
			// Extract the 'step_type' field from the stepData
			var stepType struct {
				StepType string `json:"step_type"`
			}
			if err := json.Unmarshal(stepData, &stepType); err != nil {
				return err
			}

			switch stepType.StepType {
			case schema.BlockTypePipelineStepHttp:
				var step PipelineStepHttp
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			case schema.BlockTypePipelineStepSleep:
				var step PipelineStepSleep
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			case schema.BlockTypePipelineStepEmail:
				var step PipelineStepEmail
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			case schema.BlockTypePipelineStepEcho:
				var step PipelineStepEcho
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			default:
				// Handle unrecognized step types or return an error
				return fperr.BadRequestWithMessage("Unrecognized step type: " + stepType.StepType)
			}
		}
	}

	return nil
}

func (p *Pipeline) SetAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDescription:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(nil)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse description attribute",
						Subject:  &attr.Range,
					})
					continue
				}

				valString := val.AsString()
				p.Description = &valString
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for pipeline: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
}

type PipelineOutput struct {
	Name            string         `json:"name"`
	DependsOn       []string       `json:"depends_on,omitempty"`
	Resolved        bool           `json:"resolved,omitempty"`
	Sensitive       bool           `json:"sensitive,omitempty"`
	Value           interface{}    `json:"value,omitempty"`
	UnresolvedValue hcl.Expression `json:"-"`
}

func (o *PipelineOutput) AppendDependsOn(dependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.DependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range dependsOn {
		if !existingDeps[dep] {
			o.DependsOn = append(o.DependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

type PrintablePipeline struct {
	Items interface{}
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListPipelineResponse" {
		return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	}

	lp, ok := r.(*flowpipeapiclient.ListPipelineResponse)
	if !ok {
		return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListPipelineResponse")
	}

	return lp.Items, nil
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.Pipeline)

	if !ok {
		return Table{}, fperr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []TableRow
	for _, item := range lp {
		cells := []interface{}{
			*item.Type,
			*item.Name,
			*item.Parallel,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return Table{
		Rows:    tableRows,
		Columns: p.GetColumns(),
	}, nil
}

func (PrintablePipeline) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the pipeline",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the pipeline",
		},
		{
			Name:        "PARALLEL",
			Type:        "boolean",
			Description: "Whether the pipeline is parallel",
		},
	}
}
