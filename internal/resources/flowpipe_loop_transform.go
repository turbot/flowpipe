package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type LoopTransformStep struct {
	LoopStep

	Value interface{} `json:"value,omitempty" hcl:"value,optional" cty:"value"`
}

func (l *LoopTransformStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopTransformStep, ok := other.(*LoopTransformStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopTransformStep.LoopStep) {
		return false
	}

	return reflect.DeepEqual(l.Value, otherLoopTransformStep.Value)
}

func (l *LoopTransformStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	if !helpers.IsNil(l.Value) {
		input[schema.AttributeTypeValue] = l.Value
	} else if l.UnresolvedAttributes[schema.AttributeTypeValue] != nil {
		val, err := l.UnresolvedAttributes[schema.AttributeTypeValue].Value(evalContext)
		if err != nil {
			return nil, err
		}

		if !val.IsNull() {
			goVal, err := hclhelpers.CtyToGo(val)
			if err != nil {
				return nil, err
			}
			input[schema.AttributeTypeValue] = goVal
		}
	}

	return input, nil
}

func (*LoopTransformStep) GetType() string {
	return schema.BlockTypePipelineStepTransform
}

func (s *LoopTransformStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := s.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeValue:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, s, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val == cty.NilVal {
				continue
			}

			goVal, err := hclhelpers.CtyToGo(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid attribute",
					Detail:   "Invalid attribute 'value' in the step loop block",
					Subject:  &attr.Range,
				})
				continue
			}
			s.Value = goVal

		case schema.AttributeTypeUntil:
			// already handled in SetAttributes
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Invalid attribute '" + name + "' in step loop block",
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}
