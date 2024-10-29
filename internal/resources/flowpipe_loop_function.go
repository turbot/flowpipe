package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/iancoleman/strcase"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopFunctionStep struct {
	LoopStep

	Runtime *string                 `json:"runtime,omitempty"`
	Source  *string                 `json:"source,omitempty"`
	Handler *string                 `json:"handler,omitempty"`
	Event   *map[string]interface{} `json:"event,omitempty"`
	Env     *map[string]string      `json:"env,omitempty"`
}

func (l *LoopFunctionStep) Equals(other LoopDefn) bool {
	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || l != nil && helpers.IsNil(other) {
		return false
	}

	otherLoopFunctionStep, ok := other.(*LoopFunctionStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopFunctionStep.LoopStep) {
		return false
	}

	// compare env using reflection
	if !reflect.DeepEqual(l.Env, otherLoopFunctionStep.Env) || !reflect.DeepEqual(l.Event, otherLoopFunctionStep.Event) {
		return false
	}

	return utils.PtrEqual(l.Runtime, otherLoopFunctionStep.Runtime) &&
		utils.PtrEqual(l.Handler, otherLoopFunctionStep.Handler) &&
		utils.PtrEqual(l.Source, otherLoopFunctionStep.Source)
}

func (*LoopFunctionStep) GetType() string {
	return schema.BlockTypePipelineStepFunction
}

func (l *LoopFunctionStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypeRuntime, l.Runtime)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("function", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeSource, l.Source)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeHandler, l.Handler)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = stringMapInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeEnv, l.Env)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = mapInterfaceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeEvent, l.Event)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	return result, nil
}

func (l *LoopFunctionStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeRuntime, schema.AttributeTypeSource, schema.AttributeTypeHandler:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		case schema.AttributeTypeEnv:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val == cty.NilVal {
				continue
			}

			env, err := hclhelpers.CtyToGoMapString(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid env",
					Detail:   "Invalid env in the step loop block",
					Subject:  &attr.Range,
				})
				continue
			}

			l.Env = &env

		case schema.AttributeTypeEvent:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val == cty.NilVal {
				continue
			}

			event, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid env",
					Detail:   "Invalid env in the step loop block",
					Subject:  &attr.Range,
				})
				continue
			}

			l.Event = &event

		case schema.AttributeTypeUntil:
			// already handled in SetAttributes
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Invalid attribute '" + name + "' in the step loop block",
				Subject:  &attr.Range,
			})
		}

	}
	return diags
}
