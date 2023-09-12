package modconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/options"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/terraform-components/configs"
	"github.com/zclconf/go-cty/cty"
)

func NewPipelineHcl(mod *Mod, block *hcl.Block) *Pipeline {

	pipelineFullName := block.Labels[0]

	// TODO: rethink this area, we need to be able to handle pipelines that are not in a mod
	// TODO: we're trying to integrate the pipeline & trigger functionality into the mod system, so it will look
	// TODO: like a clutch for now
	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		pipelineFullName = modName + ".pipeline." + pipelineFullName
	} else {
		pipelineFullName = "local.pipeline." + pipelineFullName
	}

	pipeline := &Pipeline{
		HclResourceImpl: HclResourceImpl{
			// The FullName is the full name of the resource, including the mod name
			FullName:        pipelineFullName,
			UnqualifiedName: "pipeline." + block.Labels[0],
		},
		// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
		PipelineName: pipelineFullName,
		Params:       map[string]*configs.Variable{},
		mod:          mod,
	}

	return pipeline
}

// Pipeline represents a "pipeline" block in an flowpipe HCL (*.fp) file
//
// Note that this Pipeline definition is different that the pipeline that is running. This definition
// contains unresolved expressions (mostly in steps), how to handle errors etc but not the actual Pipeline
// execution data.
type Pipeline struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	mod *Mod

	// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
	PipelineName string `json:"pipeline_name"`

	// Unparsed HCL body, needed so we can de-code the step HCL into the correct struct
	RawBody hcl.Body `json:"-" hcl:",remain"`

	// Unparsed JSON raw message, needed so we can unmarshall the step JSON into the correct struct
	StepsRawJson json.RawMessage `json:"-"`

	Steps []IPipelineStep `json:"steps,omitempty"`

	OutputConfig []PipelineOutput `json:"outputs,omitempty"`

	// TODO: we reduce the attributes returned by pipeline list for now, we need to decide how we want to return the data to the client
	// TODO: how do we represent the variables? They don't show up because they are stored as non serializable types for now (see UnresolvedVariables in Step)
	Params map[string]*configs.Variable `json:"-"`
}

func (p *Pipeline) GetMod() *Mod {
	return p.mod
}

// Pipeline functions
func (p *Pipeline) GetStep(stepFullyQualifiedName string) IPipelineStep {
	for i := 0; i < len(p.Steps); i++ {
		if p.Steps[i].GetFullyQualifiedName() == stepFullyQualifiedName {
			return p.Steps[i]
		}
	}
	return nil
}

func (p *Pipeline) AsCtyValue() cty.Value {
	pipelineVars := map[string]cty.Value{}
	pipelineVars[schema.LabelName] = cty.StringVal(p.Name())

	if p.Description != nil {
		pipelineVars[schema.AttributeTypeDescription] = cty.StringVal(*p.Description)
	}

	return cty.ObjectVal(pipelineVars)
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *Pipeline) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {

	var diags hcl.Diagnostics
	switch o := opts.(type) {
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (ph *Pipeline) UnmarshalJSON(data []byte) error {
	// Define an auxiliary type to decode the JSON and capture the value of the 'ISteps' field
	type Aux struct {
		PipelineName string          `json:"pipeline_name"`
		Description  *string         `json:"description,omitempty"`
		Output       *string         `json:"output,omitempty"`
		Raw          json.RawMessage `json:"-"`
		ISteps       json.RawMessage `json:"steps"`
	}

	aux := Aux{ISteps: json.RawMessage([]byte("null"))} // Provide a default value for 'ISteps' field
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Assign values to the fields of the main struct

	ph.FullName = aux.PipelineName
	ph.PipelineName = aux.PipelineName
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
			case schema.BlockTypePipelineStepQuery:
				var step PipelineStepQuery
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			case schema.BlockTypePipelineStepPipeline:
				var step PipelineStepPipeline
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				ph.Steps = append(ph.Steps, &step)
			default:
				// Handle unrecognized step types or return an error
				// return fperr.BadRequestWithMessage("Unrecognized step type: " + stepType.StepType)
				return fmt.Errorf("unrecognized step type: %s", stepType.StepType)
			}
		}
	}

	return nil
}

func (p *Pipeline) Equals(other *Pipeline) bool {

	baseEqual := p.HclResourceImpl.Equals(&other.HclResourceImpl)
	if !baseEqual {
		return false
	}

	if len(p.Steps) != len(other.Steps) {
		return false
	}

	for i := 0; i < len(p.Steps); i++ {
		if !p.Steps[i].Equals(other.Steps[i]) {
			return false
		}
	}

	if len(p.OutputConfig) != len(other.OutputConfig) {
		return false
	}

	for i := 0; i < len(p.OutputConfig); i++ {
		if !p.OutputConfig[i].Equals(&other.OutputConfig[i]) {
			return false
		}
	}

	// TODO: other checks?
	return p.FullName == other.FullName &&
		p.GetMetadata().ModFullName == other.GetMetadata().ModFullName
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

// end pipeline functions

// Pipeline HclResource interface functions

func (p *Pipeline) OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *Pipeline) setBaseProperties() {
}

// end Pipeline Hclresource interface functions

type PipelineOutput struct {
	Name            string         `json:"name"`
	DependsOn       []string       `json:"depends_on,omitempty"`
	Resolved        bool           `json:"resolved,omitempty"`
	Sensitive       bool           `json:"sensitive,omitempty"`
	Value           interface{}    `json:"value,omitempty"`
	UnresolvedValue hcl.Expression `json:"-"`
}

func (p *PipelineOutput) Equals(other *PipelineOutput) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && other == nil {
		return true
	}

	// If one of the pointers is nil while the other is not, they are not equal
	if (p == nil && other != nil) || (p != nil && other == nil) {
		return false
	}

	// Compare Name field
	if p.Name != other.Name {
		return false
	}

	// Compare DependsOn field using deep equality
	if !reflect.DeepEqual(p.DependsOn, other.DependsOn) {
		return false
	}

	// Compare Resolved field
	if p.Resolved != other.Resolved {
		return false
	}

	// Compare Sensitive field
	if p.Sensitive != other.Sensitive {
		return false
	}

	// Compare Value field using deep equality
	if !reflect.DeepEqual(p.Value, other.Value) {
		return false
	}

	// Compare UnresolvedValue field using deep equality
	if !hclhelpers.ExpressionsEqual(p.UnresolvedValue, other.UnresolvedValue) {
		return false
	}

	// All fields are equal
	return true
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
