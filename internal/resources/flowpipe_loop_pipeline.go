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

type LoopPipelineStep struct {
	LoopStep

	Args interface{} `json:"args,omitempty" hcl:"args,optional" cty:"args"`
}

func (l *LoopPipelineStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopPipelineStep, ok := other.(*LoopPipelineStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopPipelineStep.LoopStep) {
		return false
	}

	return reflect.DeepEqual(l.Args, otherLoopPipelineStep.Args)
}

func (l *LoopPipelineStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	if !helpers.IsNil(l.Args) {
		input[schema.AttributeTypeArgs] = l.Args
	} else if attr, ok := l.GetUnresolvedAttributes()[schema.AttributeTypeArgs]; ok {
		val, diags := attr.Value(evalContext)
		if len(diags) > 0 {
			return nil, error_helpers.BetterHclDiagsToError("pipeline", diags)
		}

		if val != cty.NilVal {
			goVal, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				return nil, err
			}
			input[schema.AttributeTypeArgs] = goVal
		}
	}

	return input, nil
}

func (*LoopPipelineStep) GetType() string {
	return schema.BlockTypePipelineStepPipeline
}

func (l *LoopPipelineStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	if attr, ok := hclAttributes[schema.AttributeTypeArgs]; ok {
		val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
		if len(stepDiags) > 0 {
			diags = append(diags, stepDiags...)
			return diags
		}

		if val == cty.NilVal {
			return diags
		}

		goVal, err := hclhelpers.CtyToGo(val)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Error converting cty.Value to Go type",
				Detail:   err.Error(),
				Subject:  &attr.Range,
			})
			return diags
		}

		l.Args = goVal
	}

	return diags
}
