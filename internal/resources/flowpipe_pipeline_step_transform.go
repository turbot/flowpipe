package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepTransform struct {
	PipelineStepBase
	Value any `json:"value"`
}

func (p *PipelineStepTransform) Equals(iOther PipelineStep) bool {
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || !helpers.IsNil(iOther) && p == nil {
		return false
	}

	other, ok := iOther.(*PipelineStepTransform)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if helpers.IsNil(p.Value) && !helpers.IsNil(other.Value) {
		return false
	}

	if !helpers.IsNil(p.Value) && helpers.IsNil(other.Value) {
		return false
	}

	return reflect.DeepEqual(p.Value, other.Value)
}

func (p *PipelineStepTransform) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}

func (p *PipelineStepTransform) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {
	var value any

	value, allConnectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeValue, p.Value)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	return map[string]interface{}{
		schema.AttributeTypeValue: value,
	}, allConnectionDependencies, nil
}

func (p *PipelineStepTransform) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeValue:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVal, err := hclhelpers.CtyToGo(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeValue + " attribute to interface",
						Subject:  &attr.Range,
					})
				}

				p.Value = goVal
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Transform Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}
