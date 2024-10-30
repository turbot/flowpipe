package resources

import (
	"encoding/json"
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func NewPipeline(mod *modconfig.Mod, block *hcl.Block) *Pipeline {

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
		HclResourceImpl: modconfig.HclResourceImpl{
			// The FullName is the full name of the resource, including the mod name
			FullName:        pipelineFullName,
			ShortName:       block.Labels[0],
			UnqualifiedName: "pipeline." + block.Labels[0],
			DeclRange:       block.DefRange,
			BlockType:       block.Type,
		},
		// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
		PipelineName:   pipelineFullName,
		Params:         []PipelineParam{},
		mod:            mod,
		ModFullVersion: mod.CacheKey(),
	}

	return pipeline
}

type ResourceWithParam interface {
	GetParam(paramName string) *PipelineParam
	GetParams() []PipelineParam
}

// Pipeline represents a "pipeline" block in an flowpipe HCL (*.fp) file
//
// Note that this Pipeline definition is different that the pipeline that is running. This definition
// contains unresolved expressions (mostly in steps), how to handle errors etc but not the actual Pipeline
// execution data.
type Pipeline struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	mod *modconfig.Mod

	// TODO: hack to serialise pipeline name because HclResourceImpl is not serialised
	PipelineName string `json:"pipeline_name"`
	// To be used when passing pipeline as a parameter to another pipeline, we need to know the source mod of
	// this pipeline so we can resolve the pipeline later. Name is not enough because there may be multiple
	// versions of the same mod in the current context
	ModFullVersion string `json:"mod_full_version"`

	// Unparsed HCL body, needed so we can de-code the step HCL into the correct struct
	RawBody hcl.Body `json:"-" hcl:",remain"`

	// Unparsed JSON raw message, needed so we can unmarshall the step JSON into the correct struct
	StepsRawJson json.RawMessage `json:"-"`

	Steps           []PipelineStep   `json:"steps,omitempty"`
	OutputConfig    []PipelineOutput `json:"outputs,omitempty"`
	Params          []PipelineParam  `json:"params,omitempty"`
	FileName        string           `json:"file_name"`
	StartLineNumber int              `json:"start_line_number"`
	EndLineNumber   int              `json:"end_line_number"`
}

func (p *Pipeline) GetParams() []PipelineParam {
	return p.Params
}

func (p *Pipeline) GetParam(paramName string) *PipelineParam {
	for _, param := range p.Params {
		if param.Name == paramName {
			return &param
		}
	}
	return nil
}

func (p *Pipeline) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	p.FileName = fileName
	p.StartLineNumber = startLineNumber
	p.EndLineNumber = endLineNumber
}

// func (p *Pipeline) ValidatePipelineParam(params map[string]interface{}, evalCtx *hcl.EvalContext) []error {
// 	return ValidateParams(p, params, evalCtx)
// }

// func (p *Pipeline) CoercePipelineParams(params map[string]string, evalCtx *hcl.EvalContext) (map[string]interface{}, []error) {
// 	return CoerceParams(p, params, evalCtx)
// }

// Implements modconfig.ModItem interface
func (p *Pipeline) GetMod() *modconfig.Mod {
	return p.mod
}

// Pipeline functions
func (p *Pipeline) GetStep(stepFullyQualifiedName string) PipelineStep {
	for i := 0; i < len(p.Steps); i++ {
		if p.Steps[i].GetFullyQualifiedName() == stepFullyQualifiedName {
			return p.Steps[i]
		}
	}
	return nil
}

func (p *Pipeline) CtyValue() (cty.Value, error) {
	baseCtyValue, err := p.HclResourceImpl.CtyValue()
	if err != nil {
		return cty.NilVal, err
	}

	pipelineVars := baseCtyValue.AsValueMap()
	pipelineVars[schema.LabelName] = cty.StringVal(p.Name())
	pipelineVars["mod_full_version"] = cty.StringVal(p.ModFullVersion)

	if p.Description != nil {
		pipelineVars[schema.AttributeTypeDescription] = cty.StringVal(*p.Description)
	}

	return cty.ObjectVal(pipelineVars), nil
}

// SetOptions sets the options on the pipeline (not supported)
func (p *Pipeline) SetOptions(_ options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "options are not supported on pipelines",
		Subject:  &block.DefRange,
	}}
}

func (p *Pipeline) UnmarshalJSON(data []byte) error {
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

	p.FullName = aux.PipelineName
	p.PipelineName = aux.PipelineName
	p.Description = aux.Description
	p.StepsRawJson = []byte(aux.Raw)

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
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepSleep:
				var step PipelineStepSleep
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepEmail:
				var step PipelineStepEmail
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepTransform:
				var step PipelineStepTransform
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepQuery:
				var step PipelineStepQuery
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepPipeline:
				var step PipelineStepPipeline
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}
				p.Steps = append(p.Steps, &step)

			case schema.BlockTypePipelineStepFunction:
				var step PipelineStepFunction
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepContainer:
				var step PipelineStepContainer
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepInput:
				var step PipelineStepInput
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			case schema.BlockTypePipelineStepMessage:
				var step PipelineStepMessage
				if err := json.Unmarshal(stepData, &step); err != nil {
					return err
				}

			default:
				// Handle unrecognized step types or return an error
				return perr.BadRequestWithMessage(fmt.Sprintf("unrecognized step type '%s'", stepType.StepType))

			}
		}
	}

	return nil
}

func (p *Pipeline) Equals(other *Pipeline) bool {

	if p == nil && other == nil {
		return true
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	baseEqual := p.HclResourceImpl.Equals(&other.HclResourceImpl)
	if !baseEqual {
		return false
	}

	// Order of params does not matter, but the value does
	if len(p.Params) != len(other.Params) {
		return false
	}

	// Compare param values
	for _, v := range p.Params {
		otherParam := other.GetParam(v.Name)
		if otherParam == nil {
			return false
		}

		if !v.Equals(otherParam) {
			return false
		}
	}

	// catch name change of the other param
	for _, v := range other.Params {
		pParam := p.GetParam(v.Name)
		if pParam == nil {
			return false
		}
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

	// build map for output so it's easier to lookup
	myOutput := map[string]*PipelineOutput{}
	for i, o := range p.OutputConfig {
		myOutput[o.Name] = &p.OutputConfig[i]
	}

	otherOutput := map[string]*PipelineOutput{}
	for i, o := range other.OutputConfig {
		otherOutput[o.Name] = &other.OutputConfig[i]
	}

	for k, v := range myOutput {
		if _, ok := otherOutput[k]; !ok {
			return false
		} else if !v.Equals(otherOutput[k]) {
			return false
		}
	}

	// check name changes on the other output
	for k := range otherOutput {
		if _, ok := myOutput[k]; !ok {
			return false
		}
	}

	return p.FullName == other.FullName &&
		p.GetMetadata().ModFullName == other.GetMetadata().ModFullName
}

func (p *Pipeline) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDescription:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Description = &valString
			}
		case schema.AttributeTypeTitle:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Title = &valString
			}
		case schema.AttributeTypeDocumentation:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				p.Documentation = &valString
			}
		case schema.AttributeTypeTags:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsValueMap()
				resultMap := make(map[string]string)
				for key, value := range valString {
					if value.Type() != cty.String {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Invalid type for tag value",
							Detail:   "The tag value must be a string",
							Subject:  &attr.Range,
						})
						continue
					}
					resultMap[key] = value.AsString()
				}
				p.Tags = resultMap
			}

		case schema.AttributeTypeMaxConcurrency:
			maxConcurrency, moreDiags := hclhelpers.AttributeToInt(attr, nil, false)
			if moreDiags != nil && moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			} else {
				mcInt := int(*maxConcurrency)
				p.MaxConcurrency = &mcInt
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

func (p *Pipeline) OnDecoded(*hcl.Block, modconfig.ModResourcesProvider) hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *Pipeline) setBaseProperties() {
}

// end Pipeline Hclresource interface functions

type PipelineParam struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Optional    bool              `json:"optional,omitempty"`
	Default     cty.Value         `json:"-"`
	Enum        cty.Value         `json:"-"`
	EnumGo      []any             `json:"enum"`
	Type        cty.Type          `json:"-"`
	TypeString  string            `json:"type_string"`
	Tags        map[string]string `json:"tags,omitempty"`
	Format      string            `json:"format"`
}

func (p *PipelineParam) Equals(other *PipelineParam) bool {
	if p == nil && other == nil {
		return true
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	if p.Default.Equals(other.Default) == cty.False {
		return false
	}

	if p.Enum.Equals(other.Enum) == cty.False {
		return false
	}

	return p.Name == other.Name &&
		p.Description == other.Description &&
		p.Optional == other.Optional &&
		p.Type.Equals(other.Type)
}

func (p *PipelineParam) ValidateSetting(setting cty.Value, evalCtx *hcl.EvalContext) (bool, hcl.Diagnostics, error) {
	if setting.IsNull() {
		return true, hcl.Diagnostics{}, nil
	}

	// Helper function to perform capsule type and list type validations
	validateCustomType := func() (bool, hcl.Diagnostics) {
		ctdiags := modconfig.CustomTypeValidation(setting, p.Type, nil)
		if len(ctdiags) > 0 {
			return false, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid type for param " + p.Name,
				Detail:   "The param type is not compatible with the given value",
			}}
		}

		ctdiags = p.PipelineParamCustomValueValidation(setting, evalCtx)
		return len(ctdiags) == 0, ctdiags
	}

	// Check for capsule type or list of capsule types
	if modconfig.IsCustomType(p.Type) {
		valid, diags := validateCustomType()
		return valid, diags, nil
	} else if !hclhelpers.IsValueCompatibleWithType(p.Type, setting) {
		// This is non-capsule type compatibility check
		return false, hcl.Diagnostics{}, nil
	}

	// Enum-based validation
	valid, err := hclhelpers.ValidateSettingWithEnum(setting, p.Enum)
	return valid, hcl.Diagnostics{}, err
}

func (p *PipelineParam) PipelineParamCustomValueValidation(setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	return CustomValueValidation(p.Name, setting, evalCtx)
}

func (p *PipelineParam) IsConnectionType() bool {
	return modconfig.IsConnectionType(p.Type)
}

func (p *PipelineParam) IsNotifierType() bool {
	encapsulatedGoType, nestedCapsule := hclhelpers.IsNestedCapsuleType(p.Type)
	if !nestedCapsule {
		return false
	}

	var notifierImpl *NotifierImpl

	return encapsulatedGoType.String() == reflect.TypeOf(notifierImpl).String()
}

type PipelineOutput struct {
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	DependsOn           []string       `json:"depends_on,omitempty"`
	CredentialDependsOn []string       `json:"credential_depends_on,omitempty"`
	ConnectionDependsOn []string       `json:"connection_depends_on,omitempty"`
	Resolved            bool           `json:"resolved,omitempty"`
	Value               interface{}    `json:"value,omitempty"`
	UnresolvedValue     hcl.Expression `json:"-"`
	Range               *hcl.Range     `json:"Range"`
}

func (o *PipelineOutput) Equals(other *PipelineOutput) bool {
	// If both pointers are nil, they are considered equal
	if o == nil && other == nil {
		return true
	}

	// If one of the pointers is nil while the other is not, they are not equal
	if (o == nil && other != nil) || (o != nil && other == nil) {
		return false
	}

	// Compare Name field
	if o.Name != other.Name {
		return false
	}

	if !helpers.StringSliceEqualIgnoreOrder(o.DependsOn, other.DependsOn) {
		return false
	}

	// Compare Resolved field
	if o.Resolved != other.Resolved {
		return false
	}

	// Compare Value field using deep equality
	if !reflect.DeepEqual(o.Value, other.Value) {
		return false
	}

	// Compare UnresolvedValue field using deep equality
	if !hclhelpers.ExpressionsEqual(o.UnresolvedValue, other.UnresolvedValue) {
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

func (o *PipelineOutput) AppendCredentialDependsOn(credentialDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.CredentialDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range credentialDependsOn {
		if !existingDeps[dep] {
			o.CredentialDependsOn = append(o.CredentialDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

func (o *PipelineOutput) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range o.ConnectionDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range connectionDependsOn {
		if !existingDeps[dep] {
			o.ConnectionDependsOn = append(o.ConnectionDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}
