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

type LoopHttpStep struct {
	LoopStep

	URL            *string                 `json:"url,omitempty" hcl:"url,optional" cty:"url"`
	Method         *string                 `json:"method,omitempty" hcl:"method,optional" cty:"method"`
	RequestBody    *string                 `json:"request_body,omitempty" hcl:"request_body,optional" cty:"request_body"`
	RequestHeaders *map[string]interface{} `json:"request_headers,omitempty" hcl:"request_headers,optional" cty:"request_headers"`
	CaCertPem      *string                 `json:"ca_cert_pem,omitempty" hcl:"ca_cert_pem,optional" cty:"ca_cert_pem"`
	Insecure       *bool                   `json:"insecure,omitempty" hcl:"insecure,optional" cty:"insecure"`
}

func (l *LoopHttpStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopHttpStep, ok := other.(*LoopHttpStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopHttpStep.LoopStep) {
		return false
	}

	if l.RequestHeaders == nil && otherLoopHttpStep.RequestHeaders != nil || l.RequestHeaders != nil && otherLoopHttpStep.RequestHeaders == nil {
		return false
	}

	if l.RequestHeaders != nil {
		if !reflect.DeepEqual(*l.RequestHeaders, *otherLoopHttpStep.RequestHeaders) {
			return false
		}
	}

	return utils.PtrEqual(l.URL, otherLoopHttpStep.URL) &&
		utils.PtrEqual(l.Method, otherLoopHttpStep.Method) &&
		utils.PtrEqual(l.RequestBody, otherLoopHttpStep.RequestBody) &&
		utils.PtrEqual(l.CaCertPem, otherLoopHttpStep.CaCertPem) &&
		utils.BoolPtrEqual(l.Insecure, otherLoopHttpStep.Insecure)
}

func (l *LoopHttpStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypeUrl, l.URL)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("http", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeMethod, l.Method)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("http", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeRequestBody, l.RequestBody)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("http", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeCaCertPem, l.CaCertPem)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("http", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeInsecure, l.Insecure)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("http", diags)
	}

	if l.RequestHeaders != nil {
		input[schema.AttributeTypeRequestHeaders] = *l.RequestHeaders
	} else if l.UnresolvedAttributes[schema.AttributeTypeRequestHeaders] != nil {
		attr := l.UnresolvedAttributes[schema.AttributeTypeRequestHeaders]
		val, diags := attr.Value(evalContext)
		if len(diags) > 0 {
			return nil, error_helpers.BetterHclDiagsToError("http", diags)
		}

		if val != cty.NilVal {
			requestHeader, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				return nil, error_helpers.BetterHclDiagsToError("http", hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Invalid request_headers",
						Detail:   "Invalid request_headers in the step loop block",
						Subject:  attr.Range().Ptr(),
					},
				})
			}

			input[schema.AttributeTypeRequestHeaders] = requestHeader
		}
	}

	return result, nil

}

func (*LoopHttpStep) GetType() string {
	return schema.BlockTypePipelineStepHttp
}

func (l *LoopHttpStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, "URL", true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeMethod, schema.AttributeTypeRequestBody, schema.AttributeTypeCaCertPem:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeRequestHeaders:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val == cty.NilVal {
				continue
			}

			requestHeader, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid request_headers",
					Detail:   "Invalid request_headers in the step loop block",
					Subject:  &attr.Range,
				})
				continue
			}

			l.RequestHeaders = &requestHeader

		case schema.AttributeTypeInsecure:
			stepDiags := setBoolAttributeWithResultReference(attr, evalContext, l, "Insecure", true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

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
