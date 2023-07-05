package types

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/configschema"
	"github.com/turbot/flowpipe/pipeparser/options"
	"github.com/zclconf/go-cty/cty"
)

type Input map[string]interface{}

// StepOutput is the output from a pipeline.
type StepOutput map[string]interface{}

func (o *StepOutput) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return (*o)[key]
}

type StepError struct {
	// TODO: not sure about this
	Detail fperr.ErrorModel `json:"detail"`
}

type NextStep struct {
	StepName string `json:"step_name"`
	DelayMs  int    `json:"delay_ms,omitempty"`
}

type PipelineStepError struct {
	Ignore  bool `yaml:"ignore" json:"ignore"`
	Retries int  `yaml:"retries" json:"retries"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []PipelineHcl `json:"items"`
	NextToken *string       `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID           string `json:"execution_id"`
	PipelineExecutionID   string `json:"pipeline_execution_id"`
	ParentStepExecutionID string `json:"parent_step_execution_id"`
}

type CmdPipeline struct {
	Command string `json:"command" binding:"required,oneof=run"`
}

func NewPipelineHcl(block *hcl.Block) *PipelineHcl {
	return &PipelineHcl{
		Name: block.Labels[0],
	}
}

type PipelineHcl struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty" hcl:"description,optional" cty:"description"`
	Output      *string `json:"output,omitempty"`

	// Unparsed HCL body, needed so we can de-code the step HCL into the correct struct
	RawBody hcl.Body `json:"-" hcl:",remain"`

	// Unparsed JSON raw message, needed so we can unmarshall the step JSON into the correct struct
	Raw json.RawMessage `json:"-"`

	ISteps []PipelineHclStepI `json:"steps"`
}

func (p *PipelineHcl) GetStep(stepName string) PipelineHclStepI {
	for i := 0; i < len(p.ISteps); i++ {
		if p.ISteps[i].GetName() == stepName {
			return p.ISteps[i]
		}
	}
	return nil
}

func (ph *PipelineHcl) UnmarshalJSON(data []byte) error {
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
	ph.Output = aux.Output
	ph.Raw = []byte(aux.Raw)

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
			case configschema.BlockTypePipelineStepHttp:
				var step PipelineHclStepHttp
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.ISteps = append(ph.ISteps, &step)
			case configschema.BlockTypePipelineStepSleep:
				var step PipelineHclStepSleep
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.ISteps = append(ph.ISteps, &step)
			case configschema.BlockTypePipelineStepEmail:
				var step PipelineHclStepEmail
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.ISteps = append(ph.ISteps, &step)
			default:
				// Handle unrecognized step types or return an error
				return fperr.BadRequestWithMessage("Unrecognized step type: " + stepType.StepType)
			}
		}
	}

	return nil
}

func (p *PipelineHcl) SetAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeDescription:
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

func NewPipelineStep(stepType, stepName string) PipelineHclStepI {
	switch stepType {
	case configschema.BlockTypePipelineStepHttp:
		s := PipelineHclStepHttp{}
		s.Name = stepName
		s.Type = stepType
		return &s
	case configschema.BlockTypePipelineStepSleep:
		s := PipelineHclStepSleep{}
		s.Name = stepName
		s.Type = stepType
		return &s
	case configschema.BlockTypePipelineStepEmail:
		s := PipelineHclStepEmail{}
		s.Name = stepName
		s.Type = stepType
		return &s
	default:
		return nil
	}
}

type PipelineHclStepI interface {
	GetName() string
	GetType() string
	GetInputs() map[string]interface{}
	GetDependsOn() []string
	GetFor() string
	GetError() *PipelineStepError
	SetAttributes(hcl.Attributes) hcl.Diagnostics
}

type PipelineHclStepBase struct {
	Name string `json:"name"`
	Type string `json:"step_type"`
}

func (p *PipelineHclStepBase) GetName() string {
	return p.Name
}

func (p *PipelineHclStepBase) GetType() string {
	return p.Type
}

type PipelineHclStepHttp struct {
	PipelineHclStepBase
	Url string `json:"url"`
}

func (p *PipelineHclStepHttp) GetInputs() map[string]interface{} {
	return map[string]interface{}{
		"url": p.Url,
	}
}

func (p *PipelineHclStepHttp) GetFor() string {
	return ""
}

func (p *PipelineHclStepHttp) GetDependsOn() []string {
	return []string{}
}

func (p *PipelineHclStepHttp) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepHttp) SetAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeUrl:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(nil)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse url attribute",
						Subject:  &attr.Range,
					})
					continue
				}

				valString := val.AsString()
				p.Url = valString
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for HTTP Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
}

type PipelineHclStepSleep struct {
	PipelineHclStepBase
	Duration int64 `json:"duration"`
}

func (p *PipelineHclStepSleep) GetInputs() map[string]interface{} {
	return map[string]interface{}{
		"duration": p.Duration,
	}
}

func (p *PipelineHclStepSleep) GetDependsOn() []string {
	return []string{}
}

func (p *PipelineHclStepSleep) GetFor() string {
	return ""
}

func (p *PipelineHclStepSleep) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepSleep) SetAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeDuration:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(nil)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse duration attribute",
						Subject:  &attr.Range,
					})
					continue
				}

				if !val.AsBigFloat().IsInt() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse duration attribute, not an integer",
						Subject:  &attr.Range,
					})
					continue
				}

				valInt, _ := val.AsBigFloat().Int64()
				p.Duration = valInt
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Sleep Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
}

type PipelineHclStepEmail struct {
	PipelineHclStepBase
	To string `json:"to"`
}

func (p *PipelineHclStepEmail) GetFor() string {
	return ""
}

func (p *PipelineHclStepEmail) GetDependsOn() []string {
	return []string{}
}

func (p *PipelineHclStepEmail) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepEmail) GetInputs() map[string]interface{} {
	return map[string]interface{}{
		"to": p.To,
	}
}

func (p *PipelineHclStepEmail) SetAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeTo:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(nil)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse to attribute",
						Subject:  &attr.Range,
					})
					continue
				}
				valString := val.AsString()
				p.To = valString
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Sleep Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
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

}
