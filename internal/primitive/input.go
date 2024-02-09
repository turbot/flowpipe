package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/turbot/flowpipe/templates"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/constants"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type IntegrationType string

type Input struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

type InputIntegration interface {
	PostMessage(ctx context.Context, inputType string, prompt string, options []InputIntegrationResponseOption) (*modconfig.Output, error)
}

type InputIntegrationBase struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

func NewInputIntegrationBase(input *Input) InputIntegrationBase {
	return InputIntegrationBase{
		ExecutionID:         input.ExecutionID,
		PipelineExecutionID: input.PipelineExecutionID,
		StepExecutionID:     input.StepExecutionID,
	}
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

type InputIntegrationResponseOption struct {
	Label    *string
	Value    *string
	Selected *bool
}

type InputIntegrationSlack struct {
	InputIntegrationBase
	Token         *string
	SigningSecret *string
	WebhookUrl    *string
	Channel       *string
}

func NewInputIntegrationSlack(base InputIntegrationBase) InputIntegrationSlack {
	return InputIntegrationSlack{
		InputIntegrationBase: base,
	}
}

func (ip *InputIntegrationSlack) PostMessage(ctx context.Context, inputType string, prompt string, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	// payload for callback
	payload := map[string]any{
		"execution_id":          ip.ExecutionID,
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// encodedPayload needs to be passed into block id of first block then we can extract it on receipt of Slacks payload
	encodedPayload := base64.StdEncoding.EncodeToString(jsonPayload)
	var blocks slack.Blocks
	promptBlock := slack.NewTextBlockObject(slack.PlainTextType, prompt, false, false)

	switch inputType {
	case constants.InputTypeButton:
		header := slack.NewSectionBlock(promptBlock, nil, nil, slack.SectionBlockOptionBlockID(encodedPayload))
		var buttons []slack.BlockElement
		for _, opt := range options {
			button := slack.NewButtonBlockElement("", *opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false))
			buttons = append(buttons, slack.BlockElement(button))
		}
		action := slack.NewActionBlock("", buttons...)
		blocks.BlockSet = append(blocks.BlockSet, header, action)
	case constants.InputTypeSelect:
		header := slack.NewSectionBlock(promptBlock, nil, nil, slack.SectionBlockOptionBlockID(encodedPayload))
		blockOptions := make([]*slack.OptionBlockObject, len(options))
		for i, opt := range options {
			blockOptions[i] = slack.NewOptionBlockObject(*opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false), nil)
		}
		ph := slack.NewTextBlockObject(slack.PlainTextType, "Select option", false, false)
		s := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, ph, inputType, blockOptions...)
		action := slack.NewActionBlock("action_block", s)
		blocks.BlockSet = append(blocks.BlockSet, header, action)
	case constants.InputTypeMultiSelect:
		blockOptions := make([]*slack.OptionBlockObject, len(options))
		for i, opt := range options {
			blockOptions[i] = slack.NewOptionBlockObject(*opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false), nil)
		}
		ms := slack.NewOptionsMultiSelectBlockElement(
			slack.MultiOptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, "Select options", false, false),
			inputType,
			blockOptions...)
		block := slack.NewSectionBlock(promptBlock, nil, slack.NewAccessory(ms), slack.SectionBlockOptionBlockID(encodedPayload))
		blocks.BlockSet = append(blocks.BlockSet, block)
	case constants.InputTypeText:
		textInput := slack.NewPlainTextInputBlockElement(nil, inputType)
		input := slack.NewInputBlock(encodedPayload, promptBlock, nil, textInput)
		input.DispatchAction = true // required for being able to send event
		blocks.BlockSet = append(blocks.BlockSet, input)
	default:
		return nil, perr.InternalWithMessage(fmt.Sprintf("Type %s not yet implemented for Slack Integration", inputType))
	}

	output := modconfig.Output{}
	if !helpers.IsNil(ip.Token) && !helpers.IsNil(ip.Channel) {
		var msgOption slack.MsgOption = slack.MsgOptionBlocks(blocks.BlockSet...)
		api := slack.New(*ip.Token)
		_, _, err = api.PostMessage(*ip.Channel, msgOption, slack.MsgOptionAsUser(true))
		return &output, err
	} else {
		wMsg := slack.WebhookMessage{Blocks: &blocks}
		err = slack.PostWebhook(*ip.WebhookUrl, &wMsg)
		return &output, err
	}
}

type InputIntegrationEmail struct {
	InputIntegrationBase
	Host        *string
	Port        *int64
	SecurePort  *int64
	Tls         *string
	To          []string
	From        string
	Subject     string
	User        *string
	Pass        *string
	ResponseUrl string
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

func (ip *InputIntegrationEmail) PostMessage(ctx context.Context, inputType string, prompt string, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	var err error
	host := types.SafeString(ip.Host)
	addr := fmt.Sprintf("%s:%d", host, *ip.SecurePort) // TODO: Establish approach for using correct port/secure-port
	auth := smtp.PlainAuth("", types.SafeString(ip.User), types.SafeString(ip.Pass), host)

	from := mail.Address{
		Name:    ip.From,
		Address: ip.From,
	}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = strings.Join(ip.To, ", ")
	header["Subject"] = ip.Subject
	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	templateMessage, err := parseEmailInputTemplate(ip, prompt, options)
	if err != nil {
		return nil, err
	}
	message += templateMessage

	output := modconfig.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeStartedAt] = time.Now().UTC()
	err = smtp.SendMail(addr, auth, ip.From, ip.To, []byte(message))
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

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {

	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Input must define a type")
	}

	if _, ok := i[schema.AttributeTypeType].(string); !ok {
		return perr.BadRequestWithMessage("Input type must be a string")
	}

	// TODO: validate type is one of button, text, select, multiselect, combo, multicombo
	// TODO: other validations
	// inputType := i[schema.AttributeTypeType].(string)
	//
	// switch inputType {
	// case string(IntegrationTypeSlack):
	// 	// Validate token
	// 	if i[schema.AttributeTypeToken] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a token")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeToken].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input token must be a string")
	// 	}
	//
	// 	// Validate channel
	// 	if i[schema.AttributeTypeChannel] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a channel")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeChannel].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input channel must be a string")
	// 	}
	//
	// 	// Validate the prompt
	// 	if i[schema.AttributeTypePrompt] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a prompt")
	// 	}
	// 	if _, ok := i[schema.AttributeTypePrompt].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input prompt must be a string")
	// 	}
	//
	// 	// Validate the slack type
	// 	if i[schema.AttributeTypeSlackType] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a slack type")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeSlackType].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input slack type must be a string")
	// 	}
	//
	// 	// Validate the options
	// 	var options []string
	// 	if i[schema.AttributeTypeOptions] == nil {
	// 		return perr.BadRequestWithMessage("Slack input options must define options")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeOptions].([]string); ok {
	// 		options = i[schema.AttributeTypeOptions].([]string)
	// 	}
	// 	if _, ok := i[schema.AttributeTypeOptions].([]interface{}); ok {
	// 		for _, v := range i[schema.AttributeTypeOptions].([]interface{}) {
	// 			options = append(options, v.(string))
	// 		}
	// 	}
	// 	if len(options) == 0 {
	// 		return perr.BadRequestWithMessage("Slack input options must have at least one option")
	// 	}
	// case string(IntegrationTypeEmail):
	// }

	return nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := &modconfig.Output{}

	base := NewInputIntegrationBase(ip)
	var prompt, inputType string
	var resOptions []InputIntegrationResponseOption
	if it, ok := input[schema.AttributeTypeType].(string); ok {
		inputType = it
	}
	if p, ok := input[schema.AttributeTypePrompt].(string); ok {
		prompt = p
	}

	for _, o := range input[schema.AttributeTypeOptions].([]any) {
		opt := o.(map[string]any)
		option := InputIntegrationResponseOption{}
		if l, ok := opt[schema.AttributeTypeLabel].(string); ok {
			option.Label = &l
		}
		if v, ok := opt[schema.AttributeTypeValue].(string); ok {
			option.Value = &v
			if helpers.IsNil(option.Label) {
				option.Label = &v
			}
		}
		if s, ok := opt[schema.AttributeTypeSelected].(bool); ok {
			option.Selected = &s
		}
		resOptions = append(resOptions, option)
	}

	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			for _, n := range notifies {
				notify := n.(map[string]any)
				integration := notify["integration"].(map[string]any)
				integrationType := integration["type"].(string)

				switch integrationType {
				case schema.IntegrationTypeSlack:
					s := NewInputIntegrationSlack(base)

					if channel, ok := notify[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					} else if channel, ok := integration[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					}
					if tkn, ok := integration[schema.AttributeTypeToken].(string); ok {
						s.Token = &tkn
					}
					if ss, ok := integration[schema.AttributeTypeSigningSecret].(string); ok {
						s.SigningSecret = &ss
					}
					if wu, ok := integration[schema.AttributeTypeWebhookUrl].(string); ok {
						s.WebhookUrl = &wu
					}

					// TODO: Validate?
					_, err := s.PostMessage(ctx, inputType, prompt, resOptions)
					if err != nil {
						return nil, err
					}
				case schema.IntegrationTypeWebform:
					// TODO: implement
					return nil, perr.InternalWithMessage(fmt.Sprintf("integration type %s not yet implemented", integrationType))
				case schema.IntegrationTypeEmail:
					// TODO: implement
					return nil, perr.InternalWithMessage(fmt.Sprintf("integration type %s not yet implemented", integrationType))
				default:
					return nil, perr.InternalWithMessage(fmt.Sprintf("Unsupported integration type %s", integrationType))
				}
			}
		}
	}

	return output, nil
}

func parseEmailInputTemplate(i *InputIntegrationEmail, prompt string, responseOptions []InputIntegrationResponseOption) (string, error) {
	templateFile, err := templates.HTMLTemplate("approval-template.html")
	if err != nil {
		return "", perr.InternalWithMessage("error while reading the email template")
	}
	tmpl, err := template.New("email").Parse(string(templateFile))
	if err != nil {
		return "", perr.InternalWithMessage("error while parsing the email template")
	}

	var opts []string
	for _, opt := range responseOptions {
		opts = append(opts, *opt.Value)
	}
	data := struct {
		ExecutionID         string
		PipelineExecutionID string
		StepExecutionID     string
		Options             []string
		ResponseUrl         string
		Prompt              string
	}{
		ExecutionID:         i.ExecutionID,
		PipelineExecutionID: i.PipelineExecutionID,
		StepExecutionID:     i.StepExecutionID,
		Options:             opts,
		ResponseUrl:         i.ResponseUrl,
		Prompt:              prompt,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)
	if err != nil {
		return "", perr.BadRequestWithMessage("error while executing the email template")
	}

	tempMessage := body.String()

	return tempMessage, nil
}
