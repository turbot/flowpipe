package types

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
)

// This type is used by the API to return a list of integrations.
type ListIntegrationResponse struct {
	Items     []FpIntegration `json:"items"`
	NextToken *string         `json:"next_token,omitempty"`
}

type FpIntegration struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Description     *string `json:"description,omitempty"`
	Title           *string `json:"title,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	StartLineNumber int     `json:"start_line_number,omitempty"`
	EndLineNumber   int     `json:"end_line_number,omitempty"`
	Url             *string `json:"url,omitempty"`

	// slack
	Token         *string `json:"token,omitempty"`
	SigningSecret *string `json:"signing_secret,omitempty"`
	WebhookUrl    *string `json:"webhook_url,omitempty"`
	Channel       *string `json:"channel,omitempty"`

	// email
	SmtpHost     *string  `json:"smtp_host,omitempty"`
	SmtpTls      *string  `json:"smtp_tls,omitempty"`
	SmtpPort     *int     `json:"smtp_port,omitempty"`
	SmtpsPort    *int     `json:"smtps_port,omitempty"`
	SmtpUsername *string  `json:"smtp_username,omitempty"`
	SmtpPassword *string  `json:"smtp_password,omitempty"`
	From         *string  `json:"from,omitempty"`
	To           []string `json:"to,omitempty"`
	Cc           []string `json:"cc,omitempty"`
	Bcc          []string `json:"bcc,omitempty"`
	Subject      *string  `json:"subject,omitempty"`
}

func (f FpIntegration) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var output string
	keyWidth := 10
	if f.Description != nil || f.Url != nil {
		keyWidth = 13
	}
	if f.Type == "slack" && f.SigningSecret != nil {
		keyWidth = 16
	} else if f.Type == "email" && (f.SmtpUsername != nil || f.SmtpPassword != nil) {
		keyWidth = 15
	}

	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Name:"), f.Name)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Type:"), f.Type)
	if f.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:"), *f.Title)
	}
	if f.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:"), *f.Description)
	}
	if f.Url != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Request URL:"), *f.Url)
	}

	switch f.Type {
	case "slack":
		if f.Channel != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Channel:"), *f.Channel)
		}
		if f.Token != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Token:"), *f.Token)
		}
		if f.SigningSecret != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Signing Secret:"), *f.Channel)
		}
		if f.WebhookUrl != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Webhook URL:"), *f.WebhookUrl)
		}
	case "email":
		if f.SmtpHost != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Smtp Host:"), *f.SmtpHost)
		}
		if f.SmtpPort != nil {
			output += fmt.Sprintf("%-*s%d\n", keyWidth, au.Blue("Smtp Port:"), *f.SmtpPort)
		}
		if f.SmtpsPort != nil {
			output += fmt.Sprintf("%-*s%d\n", keyWidth, au.Blue("Smtps Port:"), *f.SmtpsPort)
		}
		if f.SmtpTls != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Smtp TLS:"), *f.SmtpTls)
		}
		if f.SmtpUsername != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Smtp Username:"), *f.SmtpUsername)
		}
		if f.SmtpPassword != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Smtp Password:"), *f.SmtpPassword)
		}
		if f.From != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("From:"), *f.From)
		}
		if len(f.To) > 0 {
			output += fmt.Sprintf("%-*s\n", keyWidth, au.Blue("To:"))
			output += printItems(f.To, 2)
		}
		if len(f.Cc) > 0 {
			output += fmt.Sprintf("%-*s\n", keyWidth, au.Blue("CC:"))
			output += printItems(f.Cc, 2)
		}
		if len(f.Bcc) > 0 {
			output += fmt.Sprintf("%-*s\n", keyWidth, au.Blue("BCC:"))
			output += printItems(f.Bcc, 2)
		}
	}

	return output
}

func ListIntegrationResponseFromAPI(apiResp *flowpipeapiclient.ListIntegrationResponse) *ListIntegrationResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListIntegrationResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]FpIntegration, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpIntegrationFromAPI(apiItem)
	}
	return res
}

func FpIntegrationFromAPI(apiIntegration flowpipeapiclient.FpIntegration) FpIntegration {
	res := FpIntegration{
		Name:          typehelpers.SafeString(apiIntegration.Name),
		Type:          typehelpers.SafeString(apiIntegration.Type),
		Description:   apiIntegration.Description,
		Title:         apiIntegration.Title,
		Url:           apiIntegration.Url,
		Channel:       apiIntegration.Channel,
		Token:         apiIntegration.Token,
		SigningSecret: apiIntegration.SigningSecret,
		WebhookUrl:    apiIntegration.WebhookUrl,
		SmtpHost:      apiIntegration.SmtpHost,
		SmtpTls:       apiIntegration.SmtpTls,
		SmtpUsername:  apiIntegration.SmtpUsername,
		SmtpPassword:  apiIntegration.SmtpPassword,
		From:          apiIntegration.From,
		To:            apiIntegration.To,
		Cc:            apiIntegration.Cc,
		Bcc:           apiIntegration.Bcc,
	}
	if !helpers.IsNil(apiIntegration.SmtpPort) {
		p := int(*apiIntegration.SmtpPort)
		res.SmtpPort = &p
	}

	if !helpers.IsNil(apiIntegration.SmtpsPort) {
		p := int(*apiIntegration.SmtpsPort)
		res.SmtpsPort = &p
	}
	return res
}

func FpIntegrationFromModIntegration(integration modconfig.Integration) (*FpIntegration, error) {
	resp := &FpIntegration{
		Name:        integration.Name(),
		Type:        integration.GetIntegrationType(),
		Url:         integration.GetIntegrationImpl().Url,
		Description: integration.GetHclResourceImpl().Description,
	}

	resp.FileName = integration.GetIntegrationImpl().FileName
	resp.StartLineNumber = integration.GetIntegrationImpl().StartLineNumber
	resp.EndLineNumber = integration.GetIntegrationImpl().EndLineNumber
	redactedValue := sanitize.RedactedStr
	switch integration.GetIntegrationType() {
	case "slack":
		slack := integration.(*modconfig.SlackIntegration)
		resp.Channel = slack.Channel
		if !helpers.IsNil(slack.Token) {
			resp.Token = &redactedValue
		}
		if !helpers.IsNil(slack.WebhookUrl) {
			resp.WebhookUrl = &redactedValue
		}
		if !helpers.IsNil(slack.SigningSecret) {
			resp.SigningSecret = &redactedValue
		}
	case "email":
		email := integration.(*modconfig.EmailIntegration)
		resp.SmtpHost = email.SmtpHost
		resp.SmtpPort = email.SmtpPort
		resp.SmtpsPort = email.SmtpsPort
		resp.SmtpTls = email.SmtpTls
		resp.SmtpUsername = email.SmtpUsername
		if !helpers.IsNil(email.SmtpPassword) {
			resp.SmtpPassword = &redactedValue
		}
		resp.From = email.From
		resp.To = email.To
		resp.Cc = email.Cc
		resp.Bcc = email.Bcc
		resp.Subject = email.Subject
	}

	return resp, nil
}

func NewPrintableIntegration(resp *ListIntegrationResponse) *PrintableIntegration {
	return &PrintableIntegration{
		Items: resp.Items,
	}
}

func NewPrintableIntegrationFromSingle(input *FpIntegration) *PrintableIntegration {
	return &PrintableIntegration{
		Items: []FpIntegration{*input},
	}
}

type PrintableIntegration struct {
	Items []FpIntegration
}

func (p PrintableIntegration) GetItems() []FpIntegration {
	return p.Items
}

func (p PrintableIntegration) GetTable() (*printers.Table, error) {
	var tableRows []printers.TableRow
	for _, item := range p.Items {

		var description string
		if item.Description != nil {
			description = *item.Description
		}

		url := util.SafeDeref(item.Url)

		cells := []any{
			item.Name,
			item.Type,
			description,
			url,
		}

		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableIntegration) getColumns() (columns []string) {
	return []string{"NAME", "TYPE", "DESCRIPTION", "REQUEST URL"}
}
