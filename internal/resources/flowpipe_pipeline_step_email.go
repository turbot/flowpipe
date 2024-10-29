package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepEmail struct {
	PipelineStepBase
	To           []string `json:"to"`
	From         *string  `json:"from"`
	SmtpPassword *string  `json:"smtp_password"`
	SmtpUsername *string  `json:"smtp_username"`
	Host         *string  `json:"host"`
	Port         *int64   `json:"port"`
	SenderName   *string  `json:"sender_name"`
	Cc           []string `json:"cc"`
	Bcc          []string `json:"bcc"`
	Body         *string  `json:"body"`
	ContentType  *string  `json:"content_type"`
	Subject      *string  `json:"subject"`
}

func (p *PipelineStepEmail) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepEmail)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Use reflect.DeepEqual to compare slices and pointers
	return reflect.DeepEqual(p.To, other.To) &&
		reflect.DeepEqual(p.From, other.From) &&
		reflect.DeepEqual(p.SmtpUsername, other.SmtpUsername) &&
		reflect.DeepEqual(p.SmtpPassword, other.SmtpPassword) &&
		reflect.DeepEqual(p.Host, other.Host) &&
		reflect.DeepEqual(p.Port, other.Port) &&
		reflect.DeepEqual(p.SenderName, other.SenderName) &&
		reflect.DeepEqual(p.Cc, other.Cc) &&
		reflect.DeepEqual(p.Bcc, other.Bcc) &&
		reflect.DeepEqual(p.Body, other.Body) &&
		reflect.DeepEqual(p.ContentType, other.ContentType) &&
		reflect.DeepEqual(p.Subject, other.Subject)

}

func (p *PipelineStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}
func (p *PipelineStepEmail) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {

	results := map[string]interface{}{}
	var allConnectionDependencies []ConnectionDependency

	// to
	toValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeTo, p.To)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeTo] = toValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// from
	fromValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeFrom, p.From)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeFrom] = fromValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// smtp_username
	smtpUsernameValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSmtpUsername, p.SmtpUsername)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSmtpUsername] = smtpUsernameValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// smtp_password
	smtpPasswordValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSmtpPassword, p.SmtpPassword)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSmtpPassword] = smtpPasswordValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// host
	hostValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeHost, p.Host)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeHost] = hostValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// port
	portValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypePort, p.Port)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	if portValueInt, ok := portValue.(int); ok {
		portValue = int64(portValueInt)
	}

	results[schema.AttributeTypePort] = portValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// sender_name
	senderNameValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSenderName, p.SenderName)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSenderName] = senderNameValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// body
	bodyValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeBody, p.Body)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeBody] = bodyValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// subject
	subjectValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSubject, p.Subject)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSubject] = subjectValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// content_type
	contentTypeValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeContentType, p.ContentType)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeContentType] = contentTypeValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// cc
	ccValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeCc, p.Cc)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeCc] = ccValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// bcc
	bccValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeBcc, p.Bcc)
	if len(diags) > 0 {
		return nil, nil, error_helpers.HclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeBcc] = bccValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	return results, allConnectionDependencies, nil
}

func (p *PipelineStepEmail) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeTo:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				emailRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeTo + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.To = emailRecipients
			}

		case schema.AttributeTypeFrom:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				from, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeFrom + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.From = &from
			}

		case schema.AttributeTypeSmtpUsername:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpUsername, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpUsername + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpUsername = &smtpUsername
			}

		case schema.AttributeTypeSmtpPassword:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpPassword, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpPassword + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpPassword = &smtpPassword
			}

		case schema.AttributeTypeHost:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				host, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeHost + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Host = &host
			}

		case schema.AttributeTypePort:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				port, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to convert port into integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Port = port
			}

		case schema.AttributeTypeSenderName:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				senderName, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSenderName + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SenderName = &senderName
			}

		case schema.AttributeTypeCc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cc = ccRecipients
			}

		case schema.AttributeTypeBcc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				bccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBcc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Bcc = bccRecipients
			}

		case schema.AttributeTypeBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				body, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Body = &body
			}

		case schema.AttributeTypeContentType:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				contentType, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeContentType + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.ContentType = &contentType
			}

		case schema.AttributeTypeSubject:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				subject, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSubject + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Subject = &subject
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Email Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}
