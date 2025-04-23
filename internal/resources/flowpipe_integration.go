package resources

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type Integration interface {
	modconfig.HclResource
	modconfig.ResourceWithMetadata

	CtyValue() (cty.Value, error)
	GetIntegrationImpl() *IntegrationImpl
	GetIntegrationType() string
	MapInterface() (map[string]interface{}, error)
	SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics
	SetFileReference(fileName string, startLineNumber int, endLineNumber int)
	SetUrl(string)
	Validate() hcl.Diagnostics

	Equals(Integration) bool
}

type IntegrationImpl struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// Slack and Http has URL, Email integration does not it will be null
	Url *string `json:"url,omitempty" cty:"url" hcl:"url,optional"`

	FileName        string
	StartLineNumber int
	EndLineNumber   int
}

func (i *IntegrationImpl) SetUrl(url string) {
	i.Url = &url
}

func (i *IntegrationImpl) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	i.FileName = fileName
	i.StartLineNumber = startLineNumber
	i.EndLineNumber = endLineNumber
}

func (i *IntegrationImpl) GetIntegrationImpl() *IntegrationImpl {
	return i
}

func DefaultIntegrations() (map[string]Integration, error) {
	integrations := make(map[string]Integration)

	defaultDescription := "Default http integration"
	httpIntegration := &HttpIntegration{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName:        schema.IntegrationTypeHttp + ".default",
			ShortName:       "default",
			UnqualifiedName: schema.IntegrationTypeHttp + ".default",
			Description:     &defaultDescription,
		},
		Type: schema.IntegrationTypeHttp,
	}

	integrations[schema.IntegrationTypeHttp+".default"] = httpIntegration

	return integrations, nil
}

type SlackIntegration struct {
	modconfig.HclResourceImpl          `json:"-"`
	modconfig.ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl                    `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// slack
	Token         *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	SigningSecret *string `json:"signing_secret,omitempty" cty:"signing_secret" hcl:"signing_secret,optional"`
	WebhookUrl    *string `json:"webhook_url,omitempty" cty:"webhook_url" hcl:"webhook_url,optional"`
	Channel       *string `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
}

func (i *SlackIntegration) Equals(other Integration) bool {

	if i == nil && helpers.IsNil(other) {
		return true
	}

	if i == nil && !helpers.IsNil(other) || i != nil && helpers.IsNil(other) {
		return false
	}

	otherSlack, ok := other.(*SlackIntegration)
	if !ok {
		return false
	}

	return i.FileName == otherSlack.FileName &&
		i.StartLineNumber == otherSlack.StartLineNumber &&
		i.EndLineNumber == otherSlack.EndLineNumber &&
		((i.Token == nil && otherSlack.Token == nil) ||
			(i.Token != nil && otherSlack.Token != nil && *i.Token == *otherSlack.Token)) &&
		((i.SigningSecret == nil && otherSlack.SigningSecret == nil) ||
			(i.SigningSecret != nil && otherSlack.SigningSecret != nil && *i.SigningSecret == *otherSlack.SigningSecret)) &&
		((i.WebhookUrl == nil && otherSlack.WebhookUrl == nil) ||
			(i.WebhookUrl != nil && otherSlack.WebhookUrl != nil && *i.WebhookUrl == *otherSlack.WebhookUrl)) &&
		((i.Channel == nil && otherSlack.Channel == nil) ||
			(i.Channel != nil && otherSlack.Channel != nil && *i.Channel == *otherSlack.Channel))
}

func (i *SlackIntegration) CtyValue() (cty.Value, error) {
	iCty, err := cty_helpers.GetCtyValue(i)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := iCty.AsValueMap()
	valueMap["full_name"] = cty.StringVal(i.FullName)
	valueMap["short_name"] = cty.StringVal(i.ShortName)
	valueMap["unqualified_name"] = cty.StringVal(i.UnqualifiedName)

	if i.Title != nil {
		valueMap["title"] = cty.StringVal(*i.Title)
	}

	if i.Description != nil {
		valueMap["description"] = cty.StringVal(*i.Description)
	}

	// if i.Documentation != nil {
	// 	valueMap["documentation"] = cty.StringVal(*i.Documentation)
	// }

	return cty.ObjectVal(valueMap), nil
}

func (i *SlackIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type
	if i.Token != nil {
		res["token"] = *i.Token
	}
	if i.SigningSecret != nil {
		res["signing_secret"] = *i.SigningSecret
	}
	if i.WebhookUrl != nil {
		res["webhook_url"] = *i.WebhookUrl
	}
	if i.Channel != nil {
		res["channel"] = *i.Channel
	}

	res["full_name"] = i.FullName
	res["short_name"] = i.ShortName
	res["unqualified_name"] = i.UnqualifiedName

	if i.Title != nil {
		res["title"] = *i.Title
	}
	if i.Description != nil {
		res["description"] = *i.Description
	}

	return res, nil
}

func (i *SlackIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *SlackIntegration) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var token, webhook, signingSecret string

	// Get the token
	if i.Token != nil {
		token = *i.Token
	}

	// Get the webhook URL
	if i.WebhookUrl != nil {
		webhook = *i.WebhookUrl
	}

	// Return error if neither token nor webhook URL are not provided
	if token == "" && webhook == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  i.Name() + " requires one of the following attributes set: " + schema.AttributeTypeToken + ", " + schema.AttributeTypeWebhookUrl,
		})
	}

	// Return error if both token and webhook URL provided
	if token != "" && webhook != "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attributes " + schema.AttributeTypeToken + " and " + schema.AttributeTypeWebhookUrl + " are mutually exclusive: " + i.Name(),
		})
	}

	// Get the signing secret
	if i.SigningSecret != nil {
		signingSecret = *i.SigningSecret
	}

	// Return error if signing secret is defined when token is not provided
	if token == "" && signingSecret != "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeSigningSecret + " is only applies when attribute token is provided: " + i.Name(),
		})
	}

	return diags
}

func (i *SlackIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeToken:
			token, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.Token = token
		case schema.AttributeTypeSigningSecret:
			ss, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SigningSecret = ss
		case schema.AttributeTypeWebhookUrl:
			webhookUrl, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.WebhookUrl = webhookUrl
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Slack Integration: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

type EmailIntegration struct {
	modconfig.HclResourceImpl          `json:"-"`
	modconfig.ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl                    `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// email
	SmtpHost     *string `json:"smtp_host,omitempty" cty:"smtp_host" hcl:"smtp_host"`
	SmtpTls      *string `json:"smtp_tls,omitempty" cty:"smtp_tls" hcl:"smtp_tls,optional"`
	SmtpPort     *int    `json:"smtp_port,omitempty" cty:"smtp_port" hcl:"smtp_port,optional"`
	SmtpsPort    *int    `json:"smtps_port,omitempty" cty:"smtps_port" hcl:"smtps_port,optional"`
	SmtpUsername *string `json:"smtp_username,omitempty" cty:"smtp_username" hcl:"smtp_username,optional"`
	SmtpPassword *string `json:"smtp_password,omitempty" cty:"smtp_password" hcl:"smtp_password,optional"`

	From    *string  `json:"from,omitempty" cty:"from" hcl:"from"`
	To      []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
	Cc      []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc     []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Subject *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
}

func (i *EmailIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type
	if i.SmtpHost != nil {
		res["smtp_host"] = *i.SmtpHost
	}
	if i.SmtpTls != nil {
		res["smtp_tls"] = *i.SmtpTls
	}
	if i.SmtpPort != nil {
		res["smtp_port"] = *i.SmtpPort
	}
	if i.SmtpsPort != nil {
		res["smtps_port"] = *i.SmtpsPort
	}
	if i.SmtpUsername != nil {
		res["smtp_username"] = *i.SmtpUsername
	}
	if i.SmtpPassword != nil {
		res["smtp_password"] = *i.SmtpPassword
	}

	if i.From != nil {
		res["from"] = *i.From
	}
	if len(i.To) > 0 {
		res["to"] = i.To
	}
	if len(i.Cc) > 0 {
		res[schema.AttributeTypeCc] = i.Cc
	}
	if len(i.Bcc) > 0 {
		res["bcc"] = i.Bcc
	}

	if i.Subject != nil {
		res["subject"] = *i.Subject
	}

	res["full_name"] = i.FullName
	res["short_name"] = i.ShortName
	res["unqualified_name"] = i.UnqualifiedName

	if i.Title != nil {
		res["title"] = *i.Title
	}
	if i.Description != nil {
		res["description"] = *i.Description
	}

	return res, nil
}

func (i *EmailIntegration) Equals(other Integration) bool {

	if i == nil && helpers.IsNil(other) {
		return true
	}

	if i == nil && !helpers.IsNil(other) || i != nil && helpers.IsNil(other) {
		return false
	}

	otherEmail, ok := other.(*EmailIntegration)
	if !ok {
		return false
	}

	return i.FileName == otherEmail.FileName &&
		i.StartLineNumber == otherEmail.StartLineNumber &&
		i.EndLineNumber == otherEmail.EndLineNumber &&
		utils.PtrEqual(i.SmtpHost, otherEmail.SmtpHost) &&
		utils.PtrEqual(i.SmtpTls, otherEmail.SmtpTls) &&
		utils.PtrEqual(i.SmtpPort, otherEmail.SmtpPort) &&
		utils.PtrEqual(i.SmtpsPort, otherEmail.SmtpsPort) &&
		utils.PtrEqual(i.SmtpUsername, otherEmail.SmtpUsername) &&
		utils.PtrEqual(i.SmtpPassword, otherEmail.SmtpPassword) &&
		utils.PtrEqual(i.From, otherEmail.From) &&
		utils.PtrEqual(i.Subject, otherEmail.Subject) &&
		helpers.StringSliceEqualIgnoreOrder(i.To, otherEmail.To) &&
		helpers.StringSliceEqualIgnoreOrder(i.Cc, otherEmail.Cc) &&
		helpers.StringSliceEqualIgnoreOrder(i.Bcc, otherEmail.Bcc)
}

func (i *EmailIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *EmailIntegration) CtyValue() (cty.Value, error) {
	iCty, err := cty_helpers.GetCtyValue(i)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := iCty.AsValueMap()
	valueMap["full_name"] = cty.StringVal(i.FullName)
	valueMap["short_name"] = cty.StringVal(i.ShortName)
	valueMap["unqualified_name"] = cty.StringVal(i.UnqualifiedName)

	if i.Title != nil {
		valueMap["title"] = cty.StringVal(*i.Title)
	}

	if i.Description != nil {
		valueMap["description"] = cty.StringVal(*i.Description)
	}

	return cty.ObjectVal(valueMap), nil
}

func (i *EmailIntegration) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var from, smtpHost string

	// Get the sender info
	if i.From != nil {
		from = *i.From
	}

	// Return error if both from and smtp_host are missing
	if from == "" && smtpHost == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing required attributes " + schema.AttributeTypeFrom + ", " + schema.AttributeTypeSmtpHost + ": " + i.Name(),
		})
	}

	// Return error if from is not provided
	if from == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeFrom + " must be defined: " + i.Name(),
		})
	}

	// Get the SMTP host
	if i.SmtpHost != nil {
		smtpHost = *i.SmtpHost
	}

	// Return error if from is not provided
	if smtpHost == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeSmtpHost + " must be defined: " + i.Name(),
		})
	}

	if i.SmtpTls != nil {
		if !constants.IsValidSmtpTls(*i.SmtpTls) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute " + schema.AttributeTypeSmtpTls + " specified with invalid value " + *i.SmtpTls + ": " + i.Name(),
			})
		}
	}

	return diags
}

func (i *EmailIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSmtpHost:
			host, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpHost = host
		case schema.AttributeTypeSmtpTls:
			tls, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpTls = tls
		case schema.AttributeTypeSmtpPort:
			port, moreDiags := hclhelpers.AttributeToInt(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			portInt := int(*port)
			i.SmtpPort = &portInt
		case schema.AttributeTypeSmtpsPort:
			port, moreDiags := hclhelpers.AttributeToInt(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			portInt := int(*port)
			i.SmtpsPort = &portInt
		case schema.AttributeTypeSmtpUsername:
			uName, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpUsername = uName
		case schema.AttributeTypeSmtpPassword:
			pass, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpPassword = pass
		case schema.AttributeTypeFrom:
			from, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.From = from

		case schema.AttributeTypeTo:
			ctyVal, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			var err error
			i.To, err = hclhelpers.CtyToGoStringSlice(ctyVal, ctyVal.Type())
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attr.Name + " attribute as string slice",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
				continue
			}

		case schema.AttributeTypeCc:
			ctyVal, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			var err error
			i.Cc, err = hclhelpers.CtyToGoStringSlice(ctyVal, ctyVal.Type())
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attr.Name + " attribute as string slice",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
				continue
			}

		case schema.AttributeTypeBcc:
			ctyVal, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			var err error
			i.Bcc, err = hclhelpers.CtyToGoStringSlice(ctyVal, ctyVal.Type())
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attr.Name + " attribute as string slice",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
				continue
			}

		case schema.AttributeTypeSubject:
			subject, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.Subject = subject
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Email Integration: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func integrationFromCtyValue(val cty.Value) (Integration, error) {

	if val.IsNull() || val == cty.NilVal {
		return nil, perr.BadRequestWithMessage("Integration is required")
	}

	if !val.Type().IsMapType() && !val.Type().IsObjectType() {
		return nil, perr.BadRequestWithMessage("Invalid integration reference")
	}

	valMap := val.AsValueMap()
	integrationType := valMap["type"]

	if integrationType.IsNull() || integrationType == cty.NilVal {
		return nil, perr.BadRequestWithMessage("Integration type is required")
	}

	switch integrationType.AsString() {
	case schema.IntegrationTypeSlack:
		return SlackIntegrationFromCtyValue(val)
	case schema.IntegrationTypeEmail:
		return EmailIntegrationFromCtyValue(val)
	case schema.IntegrationTypeHttp:
		return HttpIntegrationFromCtyValue(val)
	case schema.IntegrationTypeMsTeams:
		return MsTeamsIntegrationFromCtyValue(val)
	}
	return nil, perr.BadRequestWithMessage(fmt.Sprintf("Unsupported integration type: %s", integrationType))
}

func hclResourceImplFromVal(val cty.Value) modconfig.HclResourceImpl {
	hclResourceImpl := modconfig.HclResourceImpl{}

	valueMap := val.AsValueMap()

	if !valueMap["full_name"].IsNull() && valueMap["full_name"] != cty.NilVal {
		hclResourceImpl.FullName = valueMap["full_name"].AsString()
	}

	if !valueMap["unqualified_name"].IsNull() && valueMap["unqualified_name"] != cty.NilVal {
		hclResourceImpl.UnqualifiedName = valueMap["unqualified_name"].AsString()
	}

	if !valueMap["short_name"].IsNull() && valueMap["short_name"] != cty.NilVal {
		hclResourceImpl.ShortName = valueMap["short_name"].AsString()
	}

	if !valueMap["title"].IsNull() && valueMap["title"] != cty.NilVal {
		title := valueMap["title"].AsString()
		hclResourceImpl.Title = &title
	}

	if !valueMap["description"].IsNull() && valueMap["description"] != cty.NilVal {
		description := valueMap["description"].AsString()
		hclResourceImpl.Description = &description
	}

	return hclResourceImpl
}
func SlackIntegrationFromCtyValue(val cty.Value) (*SlackIntegration, error) {
	hclResourceImpl := hclResourceImplFromVal(val)
	i := &SlackIntegration{
		HclResourceImpl: hclResourceImpl,
	}

	i.Type = val.GetAttr("type").AsString()

	valMap := val.AsValueMap()

	token := valMap["token"]
	signingSecret := valMap["signing_secret"]
	webhookUrl := valMap["webhook_url"]
	channel := valMap["channel"]

	if !token.IsNull() {
		tokenStr := token.AsString()
		i.Token = &tokenStr
	}

	if !signingSecret.IsNull() {
		signingSecretStr := signingSecret.AsString()
		i.SigningSecret = &signingSecretStr
	}

	if !webhookUrl.IsNull() {
		webhookUrlStr := webhookUrl.AsString()
		i.WebhookUrl = &webhookUrlStr
	}

	if !channel.IsNull() {
		channelStr := channel.AsString()
		i.Channel = &channelStr
	}

	return i, nil
}

func EmailIntegrationFromCtyValue(val cty.Value) (*EmailIntegration, error) {
	hclResourceImpl := hclResourceImplFromVal(val)

	i := &EmailIntegration{
		HclResourceImpl: hclResourceImpl,
	}

	i.Type = val.GetAttr("type").AsString()

	valMap := val.AsValueMap()

	smtpHost := valMap["smtp_host"]
	smtpTls := valMap["smtp_tls"]
	smtpPort := valMap["smtp_port"]
	smtpsPort := valMap["smtps_port"]
	smtpUsername := valMap["smtp_username"]
	smtpPassword := valMap["smtp_password"]
	from := valMap["from"]
	to := valMap["to"]
	cc := valMap[schema.AttributeTypeCc]
	bcc := valMap["bcc"]
	subject := valMap["subject"]

	if !smtpHost.IsNull() {
		smtpHostStr := smtpHost.AsString()
		i.SmtpHost = &smtpHostStr
	}

	if !smtpTls.IsNull() {
		smtpTlsStr := smtpTls.AsString()
		i.SmtpTls = &smtpTlsStr
	}

	if !smtpPort.IsNull() {
		smtpPortInt, _ := smtpPort.AsBigFloat().Int64()
		n := int(smtpPortInt)
		i.SmtpPort = &n
	}

	if !smtpsPort.IsNull() {
		smtpsPortInt, _ := smtpsPort.AsBigFloat().Int64()
		n := int(smtpsPortInt)
		i.SmtpsPort = &n
	}

	if !smtpUsername.IsNull() {
		smtpUsernameStr := smtpUsername.AsString()
		i.SmtpUsername = &smtpUsernameStr
	}

	if !smtpPassword.IsNull() {
		smtpPasswordStr := smtpPassword.AsString()
		i.SmtpPassword = &smtpPasswordStr
	}

	if !from.IsNull() {
		fromStr := from.AsString()
		i.From = &fromStr
	}

	var err error
	if !to.IsNull() {
		i.To, err = hclhelpers.CtyToGoStringSlice(to, to.Type())
		if err != nil {
			return nil, err
		}
	}

	if !cc.IsNull() {
		i.Cc, err = hclhelpers.CtyToGoStringSlice(cc, cc.Type())
		if err != nil {
			return nil, err
		}
	}

	if !bcc.IsNull() {
		i.Bcc, err = hclhelpers.CtyToGoStringSlice(bcc, bcc.Type())
		if err != nil {
			return nil, err
		}
	}

	if !subject.IsNull() {
		defaultSubjectStr := subject.AsString()
		i.Subject = &defaultSubjectStr
	}

	return i, nil
}

func HttpIntegrationFromCtyValue(val cty.Value) (*HttpIntegration, error) {
	hclResourceImpl := hclResourceImplFromVal(val)

	i := &HttpIntegration{
		HclResourceImpl: hclResourceImpl,
	}

	i.Type = val.GetAttr("type").AsString()

	return i, nil
}

func MsTeamsIntegrationFromCtyValue(val cty.Value) (*MsTeamsIntegration, error) {
	hclResourceImpl := hclResourceImplFromVal(val)
	i := &MsTeamsIntegration{
		HclResourceImpl: hclResourceImpl,
	}
	i.Type = val.GetAttr("type").AsString()

	valMap := val.AsValueMap()
	webhookUrl := valMap["webhook_url"]
	if !webhookUrl.IsNull() {
		webhookUrlStr := webhookUrl.AsString()
		i.WebhookUrl = &webhookUrlStr
	}

	i.IntegrationName = val.GetAttr("integration_name").AsString()

	return i, nil
}

func HclImplFromAttributes(hclResourceImpl *modconfig.HclResourceImpl, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := hcl.Diagnostics{}

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
				hclResourceImpl.Description = &valString
			}
		case schema.AttributeTypeTitle:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				hclResourceImpl.Title = &valString
			}
		case schema.AttributeTypeDocumentation:
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
					continue
				}

				valString := val.AsString()
				hclResourceImpl.Documentation = &valString
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
					resultMap[key] = value.AsString()
				}
				hclResourceImpl.Tags = resultMap
			}
		}
	}

	return diags
}

func NewIntegrationFromBlock(block *hcl.Block) Integration {
	integrationType := block.Labels[0]
	integrationName := block.Labels[1]

	integrationFullName := integrationType + "." + integrationName

	hclResourceImpl := modconfig.HclResourceImpl{
		FullName:        integrationFullName,
		UnqualifiedName: integrationFullName,
		ShortName:       integrationName,
		DeclRange:       block.DefRange,
		BlockType:       block.Type,
	}

	switch integrationType {
	case schema.IntegrationTypeSlack:
		return &SlackIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	case schema.IntegrationTypeEmail:
		return &EmailIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	case schema.IntegrationTypeHttp:
		return &HttpIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	case schema.IntegrationTypeMsTeams:
		return &MsTeamsIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
			IntegrationName: integrationFullName,
		}
	}

	return nil
}

type HttpIntegration struct {
	modconfig.HclResourceImpl          `json:"-"`
	modconfig.ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl                    `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (i *HttpIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *HttpIntegration) Equals(other Integration) bool {

	if i == nil && helpers.IsNil(other) {
		return true
	}

	if i == nil && !helpers.IsNil(other) || i != nil && helpers.IsNil(other) {
		return false
	}

	otherHttp, ok := other.(*HttpIntegration)
	if !ok {
		return false
	}

	return i.FileName == otherHttp.FileName &&
		i.StartLineNumber == otherHttp.StartLineNumber &&
		i.EndLineNumber == otherHttp.EndLineNumber
}

func (i *HttpIntegration) CtyValue() (cty.Value, error) {
	iCty, err := cty_helpers.GetCtyValue(i)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := iCty.AsValueMap()
	valueMap["full_name"] = cty.StringVal(i.FullName)
	valueMap["short_name"] = cty.StringVal(i.ShortName)
	valueMap["unqualified_name"] = cty.StringVal(i.UnqualifiedName)

	if i.Title != nil {
		valueMap["title"] = cty.StringVal(*i.Title)
	}

	if i.Description != nil {
		valueMap["description"] = cty.StringVal(*i.Description)
	}

	// if i.Documentation != nil {
	// 	valueMap["documentation"] = cty.StringVal(*i.Documentation)
	// }

	return cty.ObjectVal(valueMap), nil
}

func (i *HttpIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})

	res["type"] = i.Type

	res["full_name"] = i.FullName
	res["short_name"] = i.ShortName
	res["unqualified_name"] = i.UnqualifiedName

	if i.Title != nil {
		res["title"] = *i.Title
	}
	if i.Description != nil {
		res["description"] = *i.Description
	}

	return res, nil
}

func (i *HttpIntegration) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (i *HttpIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type MsTeamsIntegration struct {
	// base
	modconfig.HclResourceImpl          `json:"-"`
	modconfig.ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl                    `json:"-"`
	Type                               string `json:"type" cty:"type" hcl:"type,label"`
	IntegrationName                    string `json:"integration_name" cty:"integration_name"`

	// teams
	WebhookUrl *string `json:"webhook_url,omitempty" cty:"webhook_url" hcl:"webhook_url,optional"`
}

func (i *MsTeamsIntegration) CtyValue() (cty.Value, error) {
	iCty, err := cty_helpers.GetCtyValue(i)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := iCty.AsValueMap()
	valueMap["full_name"] = cty.StringVal(i.FullName)
	valueMap["short_name"] = cty.StringVal(i.ShortName)
	valueMap["unqualified_name"] = cty.StringVal(i.UnqualifiedName)

	if i.Title != nil {
		valueMap["title"] = cty.StringVal(*i.Title)
	}

	if i.Description != nil {
		valueMap["description"] = cty.StringVal(*i.Description)
	}

	valueMap["integration_name"] = cty.StringVal(i.IntegrationName)

	return cty.ObjectVal(valueMap), nil
}

func (i *MsTeamsIntegration) Equals(other Integration) bool {
	if i == nil && helpers.IsNil(other) {
		return true
	}

	if i == nil && !helpers.IsNil(other) || i != nil && helpers.IsNil(other) {
		return false
	}

	otherTeams, ok := other.(*MsTeamsIntegration)
	if !ok {
		return false
	}

	return i.FileName == otherTeams.FileName &&
		i.StartLineNumber == otherTeams.StartLineNumber &&
		i.EndLineNumber == otherTeams.EndLineNumber &&
		((i.WebhookUrl == nil && otherTeams.WebhookUrl == nil) ||
			(i.WebhookUrl != nil && otherTeams.WebhookUrl != nil && *i.WebhookUrl == *otherTeams.WebhookUrl))
}

func (i *MsTeamsIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *MsTeamsIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type

	if i.WebhookUrl != nil {
		res["webhook_url"] = *i.WebhookUrl
	}

	res["full_name"] = i.FullName
	res["short_name"] = i.ShortName
	res["unqualified_name"] = i.UnqualifiedName

	if i.Title != nil {
		res["title"] = *i.Title
	}
	if i.Description != nil {
		res["description"] = *i.Description
	}

	res["integration_name"] = i.IntegrationName

	return res, nil
}

func (i *MsTeamsIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeWebhookUrl:
			webhookUrl, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.WebhookUrl = webhookUrl
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for msteams Integration: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func (i *MsTeamsIntegration) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var whUrl string
	if i.WebhookUrl != nil {
		whUrl = *i.WebhookUrl
	}
	if whUrl == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeWebhookUrl + " must be defined: " + i.Name(),
			Subject:  &i.DeclRange,
		})
	}

	return diags
}
