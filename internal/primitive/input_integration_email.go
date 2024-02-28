package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/slack-go/slack"
	"github.com/turbot/flowpipe/templates"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"

	kitTypes "github.com/turbot/go-kit/types"
)

type InputIntegrationEmailMessage interface {
	Message() (string, error)
}

type InputIntegrationEmail struct {
	InputIntegrationBase

	Host       *string
	Port       *int64
	SecurePort *int64
	Tls        *string
	To         []string
	Cc         []string
	Bcc        []string
	From       string
	Subject    string
	User       *string
	Pass       *string
	FormUrl    string
}

func NewInputIntegrationEmail(base InputIntegrationBase) InputIntegrationEmail {
	return InputIntegrationEmail{
		InputIntegrationBase: base,
	}
}

func (ip *InputIntegrationEmail) ValidateInputIntegrationEmail(ctx context.Context, i modconfig.Input) error {

	// Validate sender's information
	if i[schema.AttributeTypeFrom] == nil {
		return perr.BadRequestWithMessage("Email input must define from")
	}
	if i[schema.AttributeTypeSmtpPassword] == nil {
		return perr.BadRequestWithMessage("Email input must define sender_credential")
	}
	if i[schema.AttributeTypeHost] == nil {
		return perr.BadRequestWithMessage("Email input must define a SMTP host")
	}
	if i[schema.AttributeTypePort] == nil {
		return perr.BadRequestWithMessage("Email input must define a port")
	}

	// Validate the port input
	if i[schema.AttributeTypePort] != nil {
		var port int64
		switch data := i[schema.AttributeTypePort].(type) {
		case float64:
			port = int64(data)
		case int64:
			port = data
		default:
			return perr.BadRequestWithMessage("port must be a number")
		}

		portInString := strconv.FormatInt(port, 10)
		match, err := regexp.MatchString("^((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4}))$", portInString)
		if err != nil {
			return perr.BadRequestWithMessage("error while validating the port")
		}

		if !match {
			return perr.BadRequestWithMessage(fmt.Sprintf("%s is not a valid port", portInString))
		}
	}

	if i[schema.AttributeTypeSenderName] != nil {
		if _, ok := i[schema.AttributeTypeSenderName].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'sender_name' must be a string")
		}
	}

	// Validate the recipients
	if i[schema.AttributeTypeTo] == nil {
		return perr.BadRequestWithMessage("Email input must define to")
	}
	if _, ok := i[schema.AttributeTypeTo].([]string); !ok {
		// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
		// Check if the input is an interface slice, and the elements are strings
		if data, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
			for _, v := range data {
				if _, ok := v.(string); !ok {
					return perr.BadRequestWithMessage("Email attribute 'to' must have elements of type string")
				}
			}
			return nil
		}

		return perr.BadRequestWithMessage("Email attribute 'to' must be an array")
	}

	var recipients []string
	if _, ok := i[schema.AttributeTypeTo].([]string); ok {
		recipients = i[schema.AttributeTypeTo].([]string)
	}

	if _, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
		for _, v := range i[schema.AttributeTypeTo].([]interface{}) {
			recipients = append(recipients, v.(string))
		}
	}

	if len(recipients) == 0 {
		return perr.BadRequestWithMessage("Recipients must not be empty")
	}

	// Validate the Cc recipients
	if i[schema.AttributeTypeCc] != nil {
		if _, ok := i[schema.AttributeTypeCc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeCc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'cc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'cc' must be an array")
		}
	}

	// Validate the Bcc recipients
	if i[schema.AttributeTypeBcc] != nil {
		if _, ok := i[schema.AttributeTypeBcc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeBcc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'bcc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'bcc' must be an array")
		}
	}

	// Validate the email body
	if i[schema.AttributeTypeBody] != nil {
		if _, ok := i[schema.AttributeTypeBody].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'body' must be a string")
		}
	}

	if i[schema.AttributeTypeContentType] != nil {
		if _, ok := i[schema.AttributeTypeContentType].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'content_type' must be a string")
		}
	}

	// validate the subject
	if i[schema.AttributeTypeSubject] != nil {
		if _, ok := i[schema.AttributeTypeSubject].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'subject' must be a string")
		}
	}

	return nil
}

func (ip *InputIntegrationEmail) PostMessage(ctx context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	var err error
	host := kitTypes.SafeString(ip.Host)
	addr := fmt.Sprintf("%s:%d", host, *ip.SecurePort) // TODO: Establish approach for using correct port/secure-port
	auth := smtp.PlainAuth("", kitTypes.SafeString(ip.User), kitTypes.SafeString(ip.Pass), host)

	output := modconfig.Output{
		Data: map[string]interface{}{},
	}

	message, err := mc.EmailMessage(ip, options)
	if err != nil {
		return nil, perr.InternalWithMessage(fmt.Sprintf("unable to create email message: %s", err.Error()))
	}

	recipients := ip.To
	if len(ip.Cc) > 0 {
		recipients = append(recipients, ip.Cc...)
	}
	if len(ip.Bcc) > 0 {
		recipients = append(recipients, ip.Bcc...)
	}

	output.Data[schema.AttributeTypeStartedAt] = time.Now().UTC()
	err = smtp.SendMail(addr, auth, ip.From, recipients, []byte(message))
	output.Data[schema.AttributeTypeFinishedAt] = time.Now().UTC()
	if err != nil {
		var smtpError *textproto.Error
		if !errors.As(err, &smtpError) {
			return nil, perr.InternalWithMessage(fmt.Sprintf("unable to send email: %s", err.Error()))
		}
		switch {
		case smtpError.Code >= 400 && smtpError.Code <= 499:
			output.Errors = []modconfig.StepError{
				{
					Error: perr.BadRequestWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg)),
				},
			}
		case smtpError.Code >= 500 && smtpError.Code <= 599:
			output.Errors = []modconfig.StepError{
				{
					Error: perr.ServiceUnavailableWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg)),
				},
			}
		default:
			output.Errors = []modconfig.StepError{
				{
					Error: perr.InternalWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg)),
				},
			}
		}
	}

	return &output, nil
}

type InputStepMessageCreator struct {
	Prompt    string
	InputType string
	StepName  string
}

func (icm *InputStepMessageCreator) SlackMessage(ip *InputIntegrationSlack, options []InputIntegrationResponseOption) (slack.Blocks, error) {
	var blocks slack.Blocks

	// payload for callback
	payload := map[string]any{
		"execution_id":          ip.ExecutionID,
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return blocks, err
	}
	// encodedPayload needs to be passed into block id of first block then we can extract it on receipt of Slacks payload
	encodedPayload := base64.StdEncoding.EncodeToString(jsonPayload)

	promptBlock := slack.NewTextBlockObject(slack.PlainTextType, icm.Prompt, false, false)
	boldPromptBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*%s*", icm.Prompt), false, false)

	switch icm.InputType {
	case constants.InputTypeButton:
		header := slack.NewSectionBlock(boldPromptBlock, nil, nil, slack.SectionBlockOptionBlockID(encodedPayload))
		var buttons []slack.BlockElement
		for i, opt := range options {
			button := slack.NewButtonBlockElement(fmt.Sprintf("finished_%d", i), *opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false))
			if !helpers.IsNil(opt.Style) {
				switch *opt.Style {
				case constants.InputStyleOk:
					button.Style = slack.StylePrimary
				case constants.InputStyleAlert:
					button.Style = slack.StyleDanger
				}
			}
			buttons = append(buttons, slack.BlockElement(button))
		}
		action := slack.NewActionBlock("action_block", buttons...)
		blocks.BlockSet = append(blocks.BlockSet, header, action)
	case constants.InputTypeSelect:
		blockOptions := make([]*slack.OptionBlockObject, len(options))
		for i, opt := range options {
			blockOptions[i] = slack.NewOptionBlockObject(*opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false), nil)
		}
		ph := slack.NewTextBlockObject(slack.PlainTextType, "Select option", false, false)
		s := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, ph, "finished", blockOptions...)
		input := slack.NewInputBlock(encodedPayload, promptBlock, nil, s)
		blocks.BlockSet = append(blocks.BlockSet, input)
	case constants.InputTypeMultiSelect:
		blockOptions := make([]*slack.OptionBlockObject, len(options))
		for i, opt := range options {
			blockOptions[i] = slack.NewOptionBlockObject(*opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false), nil)
		}
		ms := slack.NewOptionsMultiSelectBlockElement(
			slack.MultiOptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, "Select options", false, false), "not_finished", blockOptions...)
		btn := slack.NewButtonBlockElement("finished", "submit", slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false))
		input := slack.NewInputBlock(encodedPayload, promptBlock, nil, ms)
		action := slack.NewActionBlock("action_block", btn)
		blocks.BlockSet = append(blocks.BlockSet, input, action)
	case constants.InputTypeText:
		textInput := slack.NewPlainTextInputBlockElement(nil, "finished")
		input := slack.NewInputBlock(encodedPayload, promptBlock, nil, textInput)
		input.DispatchAction = true // required for being able to send event
		blocks.BlockSet = append(blocks.BlockSet, input)
	default:
		return blocks, perr.InternalWithMessage(fmt.Sprintf("Type %s not yet implemented for Slack Integration", icm.InputType))
	}

	return blocks, nil
}

func (icm *InputStepMessageCreator) EmailMessage(iim *InputIntegrationEmail, options []InputIntegrationResponseOption) (string, error) {

	header := make(map[string]string)
	header["From"] = iim.From
	header["To"] = strings.Join(iim.To, ", ")
	if len(iim.Cc) > 0 {
		header["Cc"] = strings.Join(iim.Cc, ", ")
	}
	header["Subject"] = iim.Subject
	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	var data any
	templateFileName := "input-form-link.html"
	switch icm.InputType {
	case "button":
		stepName := icm.StepName
		templateFileName = "input-form-buttons.html"
		data = struct {
			Prompt   string
			Options  []InputIntegrationResponseOption
			StepName string
			FormUrl  string
		}{
			Prompt:   icm.Prompt,
			Options:  options,
			StepName: stepName,
			FormUrl:  iim.FormUrl,
		}
	default:
		data = struct {
			FormUrl string
			Prompt  string
		}{
			FormUrl: iim.FormUrl,
			Prompt:  icm.Prompt,
		}
	}

	templateMessage, err := parseEmailInputTemplate(templateFileName, data)
	if err != nil {
		return "", err
	}
	message += templateMessage

	return message, nil

}

func parseEmailInputTemplate(templateFileName string, data any) (string, error) {
	funcs := template.FuncMap{
		"mod": func(a, b int) int { return a % b },
		"sub": func(a, b int) int { return a - b },
		"getColor": func(s *string) string {
			if s != nil {
				switch *s {
				case "ok":
					return "#379634"
				case "alert":
					return "#b32128"
				default:
					return "#036"
				}
			}
			return "#036"
		},
	}
	templateFile, err := templates.HTMLTemplate(templateFileName)
	if err != nil {
		return "", perr.InternalWithMessage("error while reading the email template")
	}
	tmpl, err := template.New("email").Funcs(funcs).Parse(string(templateFile))
	if err != nil {
		return "", perr.InternalWithMessage("error while parsing the email template")
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)
	if err != nil {
		return "", perr.BadRequestWithMessage("error while executing the email template")
	}

	tempMessage := body.String()

	return tempMessage, nil
}
