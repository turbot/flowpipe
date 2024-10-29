package resources

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/iancoleman/strcase"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopInputStep struct {
	LoopStep

	Notifier *NotifierImpl `json:"notifier" cty:"-" hcl:"-"`

	Prompt  *string   `json:"prompt" cty:"prompt" hcl:"prompt,optional"`
	Cc      *[]string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc     *[]string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel *string   `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Subject *string   `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	To      *[]string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}

func (l *LoopInputStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || l != nil && helpers.IsNil(other) {
		return false
	}

	otherLoopInputStep, ok := other.(*LoopInputStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopInputStep.LoopStep) {
		return false
	}

	if l.Cc == nil && otherLoopInputStep.Cc != nil || l.Cc != nil && otherLoopInputStep.Cc == nil {
		return false
	} else if l.Cc != nil {
		if slices.Compare(*l.Cc, *otherLoopInputStep.Cc) != 0 {
			return false
		}
	}

	if l.Bcc == nil && otherLoopInputStep.Bcc != nil || l.Bcc != nil && otherLoopInputStep.Bcc == nil {
		return false
	} else if l.Bcc != nil {
		if slices.Compare(*l.Bcc, *otherLoopInputStep.Bcc) != 0 {
			return false
		}
	}

	if l.To == nil && otherLoopInputStep.To != nil || l.To != nil && otherLoopInputStep.To == nil {
		return false
	} else if l.To != nil {
		if slices.Compare(*l.To, *otherLoopInputStep.To) != 0 {
			return false
		}
	}

	return utils.PtrEqual(l.Prompt, otherLoopInputStep.Prompt) &&
		utils.PtrEqual(l.Channel, otherLoopInputStep.Channel) &&
		utils.PtrEqual(l.Subject, otherLoopInputStep.Subject)
}

func (s *LoopInputStep) GetType() string {
	return schema.BlockTypeInput
}

func (l *LoopInputStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypePrompt, l.Prompt)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeChannel, l.Channel)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeSubject, l.Subject)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	result, diags = stringSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeCc, l.Cc)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	result, diags = stringSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeBcc, l.Bcc)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	result, diags = stringSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeTo, l.To)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("input", diags)
	}

	if l.Notifier != nil {
		input[schema.AttributeTypeNotifier] = *l.Notifier
	} else if attr, ok := l.GetUnresolvedAttributes()[schema.AttributeTypeNotifier]; ok {
		val, diags := attr.Value(evalContext)
		if len(diags) > 0 {
			return nil, error_helpers.BetterHclDiagsToError("input", diags)
		}

		if val != cty.NilVal {
			ntfy, err := ctyValueToPipelineStepNotifierValueMap(val)
			if err != nil {
				return nil, err
			}
			input[schema.AttributeTypeNotifier] = ntfy
		}
	}

	return result, nil
}

func (l *LoopInputStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypePrompt, schema.AttributeTypeChannel, schema.AttributeTypeSubject:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeCc, schema.AttributeTypeBcc, schema.AttributeTypeTo:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringSliceAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeNotifier:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ntfy, err := ctyValueToPipelineStepNotifierValueMap(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeNotifier + " attribute to notifier",
						Detail:   err.Error(),
						Subject:  &attr.Range,
					})
				}
				l.Notifier = &ntfy
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
