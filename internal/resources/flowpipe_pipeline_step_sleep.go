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

type PipelineStepSleep struct {
	PipelineStepBase
	Duration interface{} `json:"duration"`
}

func (p *PipelineStepSleep) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || !helpers.IsNil(iOther) && p == nil {
		return false
	}

	other, ok := iOther.(*PipelineStepSleep)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return reflect.DeepEqual(p.Duration, other.Duration)
}

func (p *PipelineStepSleep) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}

func (p *PipelineStepSleep) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {
	var durationInput any
	var connectionDependencies []ConnectionDependency

	durationInput, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeDuration, p.Duration)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	return map[string]interface{}{
		schema.AttributeTypeDuration: durationInput,
	}, connectionDependencies, nil
}

func (p *PipelineStepSleep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				duration, err := hclhelpers.CtyToGo(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeDuration + "' attribute to interface",
						Subject:  &attr.Range,
					})
				}
				p.Duration = duration
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for " + schema.BlockTypePipelineStepSleep + " Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepSleep) Validate() hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	if p.Duration != nil {
		switch p.Duration.(type) {
		case string, int:
			// valid duration
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Value of the attribute '" + schema.AttributeTypeDuration + "' must be a string or a whole number: " + p.GetFullyQualifiedName(),
				Subject:  p.Range,
			})
		}
	}

	return diags
}
