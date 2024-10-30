package resources

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type LoopEmailStep struct {
	LoopStep

	To               *[]string `json:"to,omitempty" hcl:"to,optional" cty:"to"`
	From             *string   `json:"from,omitempty" hcl:"from,optional" cty:"from"`
	SenderCredential *string   `json:"sender_credential,omitempty" hcl:"sender_credential,optional" cty:"sender_credential"`
	Host             *string   `json:"host,omitempty" hcl:"host,optional" cty:"host"`
	Port             *int64    `json:"port,omitempty" hcl:"port,optional" cty:"port"`
	SenderName       *string   `json:"sender_name,omitempty" hcl:"sender_name,optional" cty:"sender_name"`
	Cc               *[]string `json:"cc,omitempty" hcl:"cc,optional" cty:"cc"`
	Bcc              *[]string `json:"bcc,omitempty" hcl:"bcc,optional" cty:"bcc"`
	Body             *string   `json:"body,omitempty" hcl:"body,optional" cty:"body"`
	ContentType      *string   `json:"content_type,omitempty" hcl:"content_type,optional" cty:"content_type"`
	Subject          *string   `json:"subject,omitempty" hcl:"subject,optional" cty:"subject"`
}

func (l *LoopEmailStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopEmailStep, ok := other.(*LoopEmailStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopEmailStep.LoopStep) {
		return false
	}

	if l.To == nil && otherLoopEmailStep.To != nil || l.To != nil && otherLoopEmailStep.To == nil {
		return false
	}

	if l.To != nil && !helpers.StringSliceEqualIgnoreOrder(*l.To, *otherLoopEmailStep.To) {
		return false
	}

	if l.Cc == nil && otherLoopEmailStep.Cc != nil || l.Cc != nil && otherLoopEmailStep.Cc == nil {
		return false
	}

	if l.Cc != nil && !helpers.StringSliceEqualIgnoreOrder(*l.Cc, *otherLoopEmailStep.Cc) {
		return false
	}

	if l.Bcc == nil && otherLoopEmailStep.Bcc != nil || l.Bcc != nil && otherLoopEmailStep.Bcc == nil {
		return false
	}

	if l.Bcc != nil && !helpers.StringSliceEqualIgnoreOrder(*l.Bcc, *otherLoopEmailStep.Bcc) {
		return false
	}

	return utils.PtrEqual(l.From, otherLoopEmailStep.From) &&
		utils.PtrEqual(l.SenderCredential, otherLoopEmailStep.SenderCredential) &&
		utils.PtrEqual(l.Host, otherLoopEmailStep.Host) &&
		utils.PtrEqual(l.Port, otherLoopEmailStep.Port) &&
		utils.PtrEqual(l.SenderName, otherLoopEmailStep.SenderName) &&
		utils.PtrEqual(l.Body, otherLoopEmailStep.Body) &&
		utils.PtrEqual(l.ContentType, otherLoopEmailStep.ContentType) &&
		utils.PtrEqual(l.Subject, otherLoopEmailStep.Subject)
}

func (l *LoopEmailStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {
	if l.To != nil {
		input["to"] = *l.To
	}
	if l.From != nil {
		input["from"] = *l.From
	}
	if l.SenderCredential != nil {
		input["sender_credential"] = *l.SenderCredential
	}
	if l.Host != nil {
		input["host"] = *l.Host
	}
	if l.Port != nil {
		input["port"] = *l.Port
	}
	if l.SenderName != nil {
		input["sender_name"] = *l.SenderName
	}
	if l.Cc != nil {
		input[schema.AttributeTypeCc] = *l.Cc
	}
	if l.Bcc != nil {
		input["bcc"] = *l.Bcc
	}
	if l.Body != nil {
		input["body"] = *l.Body
	}
	if l.ContentType != nil {
		input["content_type"] = *l.ContentType
	}
	if l.Subject != nil {
		input["subject"] = *l.Subject
	}
	return input, nil
}

func (*LoopEmailStep) GetType() string {
	return schema.BlockTypePipelineStepEmail
}

func (l *LoopEmailStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}
