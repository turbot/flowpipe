package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atc0005/go-teams-notify/v2/messagecard"
	"github.com/charmbracelet/huh"
	"github.com/slack-go/slack"
	fconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/go-kit/helpers"
	kitTypes "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type Input struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
	PipelineName        string
	StepName            string
}

func NewInputPrimitive(executionId, pipelineExecutionId, stepExecutionId, pipelineName, stepName string) *Input {
	return &Input{
		ExecutionID:         executionId,
		PipelineExecutionID: pipelineExecutionId,
		StepExecutionID:     stepExecutionId,
		PipelineName:        pipelineName,
		StepName:            stepName,
	}
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
	Label    *string `json:"label,omitempty"`
	Value    *string `json:"value,omitempty"`
	Selected *bool   `json:"selected,omitempty"`
	Style    *string `json:"style,omitempty"`
}

func validateInputStepInput(ctx context.Context, i modconfig.Input) error {
	// validate type
	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Input must define a type")
	}
	inputType, inputTypeIsString := i[schema.AttributeTypeType].(string)
	if !inputTypeIsString {
		return perr.BadRequestWithMessage("Input type must be a string")
	}
	if !constants.IsValidInputType(inputType) {
		return perr.BadRequestWithMessage(fmt.Sprintf("Input type '%s' is not supported", inputType))
	}

	// validate options
	switch inputType {
	case constants.InputTypeText:
		// text type doesn't require options, but don't fail if we have them, just ignore
	default:
		// ensure has at least 1 option
		options, hasOpts := i[schema.AttributeTypeOptions].([]any)
		if !hasOpts || len(options) == 0 {
			return perr.BadRequestWithMessage(fmt.Sprintf("Input type '%s' requires options, no options were defined", inputType))
		}

		// ensure all options have a value
		for idx, opt := range options {
			option := opt.(map[string]any)
			if helpers.IsNil(option[schema.AttributeTypeValue]) {
				return perr.BadRequestWithMessage(fmt.Sprintf("option %d has no value specified", idx))
			}
		}
	}

	return nil
}

func parseOptionsFromInput(i modconfig.Input) []InputIntegrationResponseOption {
	var resOptions []InputIntegrationResponseOption

	if options, hasOptions := i[schema.AttributeTypeOptions].([]any); hasOptions {
		for _, op := range options {
			opt := op.(map[string]any)
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
			if s, ok := opt[schema.AttributeTypeStyle].(string); ok {
				option.Style = &s
			}
			resOptions = append(resOptions, option)
		}
	}

	return resOptions
}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {
	err := validateInputStepInput(ctx, i)
	if err != nil {
		return err // will already be perr
	}

	err = ip.validateInputNotifier(i)
	if err != nil {
		return err // will already be perr
	}

	return nil
}

func (ip *Input) validateInputNotifier(i modconfig.Input) error {
	notifier := i[schema.AttributeTypeNotifier].(map[string]any)
	notifies := notifier[schema.AttributeTypeNotifies].([]any)
	for _, n := range notifies {
		notify := n.(map[string]any)
		integration := notify["integration"].(map[string]any)
		integrationType := integration["type"].(string)

		switch integrationType {
		case schema.IntegrationTypeHttp:
			// no additional validations required
		case schema.IntegrationTypeSlack:
			// if using token, we need to specify channel, webhook_url approach has a bound channel
			if integration[schema.AttributeTypeToken] != nil {
				if _, stepChannel := i[schema.AttributeTypeChannel].(string); !stepChannel {
					if _, notifyChannel := notify[schema.AttributeTypeChannel].(string); !notifyChannel {
						if _, integrationChannel := integration[schema.AttributeTypeChannel].(string); !integrationChannel {
							return perr.BadRequestWithMessage("slack notifications require a channel when using token auth, channel was not set")
						}
					}
				}
			}
		case schema.IntegrationTypeEmail:
			// ensure we have recipients, these can be to, cc or bcc but as optional at each layer need to ensure we have a target
			var recipients []string

			if to, ok := i[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			} else if to, ok := notify[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			} else if to, ok := integration[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			}

			if cc, ok := i[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			} else if cc, ok := notify[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			} else if cc, ok := integration[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			}

			if bcc, ok := i[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			} else if bcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			} else if bcc, ok := integration[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			}

			if len(recipients) == 0 {
				return perr.BadRequestWithMessage("email notifications require recipients; one of 'to', 'cc' or 'bcc' need to be set")
			}
		case schema.IntegrationTypeMsTeams:
			// no additional validations required now as >4 options on button should render as select instead of error
		}
	}

	return nil
}

func (ip *Input) execute(ctx context.Context, input modconfig.Input, mc MessageCreator) (*modconfig.Output, error) {
	output := &modconfig.Output{}

	resOptions := parseOptionsFromInput(input)

	if o.IsServerMode || (os.Getenv("RUN_MODE") == "TEST_ES") {
		extNotifySent, nErrors := ip.sendNotifications(ctx, input, mc, resOptions)

		switch {
		case !extNotifySent && len(nErrors) == 0: // no integrations or only http integrations
			return output, nil
		case extNotifySent && len(nErrors) == 0: // all notifications sent
			return output, nil
		case extNotifySent && len(nErrors) > 0: // some external notifications sent, some failed...
			// TODO: Figure out how to get this into the pipeline run output
			if o.IsServerMode {
				sp := types.NewServerOutputPrefixWithExecId(time.Now().UTC(), "pipeline", &ip.ExecutionID)
				for _, ne := range nErrors {
					o.RenderServerOutput(ctx, types.NewServerOutputError(sp, "unable to send notification", ne))
				}
			}
			return output, nil
		case !extNotifySent && len(nErrors) > 0: // all external notifications failed
			var detail string
			for _, ne := range nErrors {
				if e, ok := ne.(perr.ErrorModel); ok {
					detail += fmt.Sprintf("%s\n", e.Detail)
				} else if e, ok := ne.(slack.StatusCodeError); ok {
					detail += fmt.Sprintf("%s\n", e.Error())
				} else {
					detail += fmt.Sprintf("%s\n", ne.Error())
				}
			}
			return nil, perr.InternalWithMessage(fmt.Sprintf("all %d notifications failed:\n%s", len(nErrors), detail))
		}

		return output, nil
	}

	return ip.consoleIntegration(ctx, input, mc, resOptions)
}

func (ip *Input) sendNotifications(ctx context.Context, input modconfig.Input, mc MessageCreator, opts []InputIntegrationResponseOption) (bool, []error) {
	base := NewInputIntegrationBase(ip)
	externalNotificationSent := false
	var notificationErrors []error
	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			for _, n := range notifies {
				notify := n.(map[string]any)
				integration := notify["integration"].(map[string]any)
				integrationType := integration["type"].(string)

				switch integrationType {
				case schema.IntegrationTypeSlack:
					s := NewInputIntegrationSlack(base)

					// Three ways to set the channel, in order of precedence
					if channel, ok := input[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					} else if channel, ok := notify[schema.AttributeTypeChannel].(string); ok {
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

					_, err := s.PostMessage(ctx, mc, opts)
					if err != nil {
						notificationErrors = append(notificationErrors, err)
					} else {
						externalNotificationSent = true
					}
				case schema.IntegrationTypeHttp:
					// No output needs to be rendered here for HTTP step. The console output is rendered by the Event printer, it does the right thing there too.
				case schema.IntegrationTypeEmail:
					e := NewInputIntegrationEmail(base)

					if formUrl, ok := input[fconstants.FormUrl].(string); ok {
						e.FormUrl = formUrl
					}

					if host, ok := integration[schema.AttributeTypeSmtpHost].(string); ok {
						e.Host = &host
					}
					if port, ok := integration[schema.AttributeTypeSmtpPort].(int64); ok {
						e.Port = &port
					} else if port, ok := integration[schema.AttributeTypeSmtpPort].(float64); ok {
						intPort := int64(port)
						e.Port = &intPort
					}
					if sPort, ok := integration[schema.AttributeTypeSmtpsPort].(int64); ok {
						e.SecurePort = &sPort
					} else if sPort, ok := integration[schema.AttributeTypeSmtpsPort].(float64); ok {
						intPort := int64(sPort)
						e.SecurePort = &intPort
					}
					if tls, ok := integration[schema.AttributeTypeSmtpTls].(string); ok {
						e.Tls = &tls
					}
					if from, ok := integration[schema.AttributeTypeFrom].(string); ok {
						e.From = from
					}

					if to, ok := input[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					} else if to, ok := notify[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					} else if to, ok := integration[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					}

					if cc, ok := input[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					} else if cc, ok := notify[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					} else if cc, ok := integration[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					}

					if bcc, ok := input[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					} else if bcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					} else if bcc, ok := integration[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					}

					if sub, ok := input[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					} else if sub, ok := notify[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					} else if sub, ok := integration[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					}

					if u, ok := integration[schema.AttributeTypeSmtpUsername].(string); ok {
						e.User = &u
					}
					if p, ok := integration[schema.AttributeTypeSmtpPassword].(string); ok {
						e.Pass = &p
					}

					_, err := e.PostMessage(ctx, mc, opts)
					if err != nil {
						notificationErrors = append(notificationErrors, err)
					} else {
						externalNotificationSent = true
					}

				case schema.IntegrationTypeMsTeams:
					integrationName := integration["integration_name"].(string)
					t := NewInputIntegrationMsTeams(base, integrationName)
					if wu, ok := integration[schema.AttributeTypeWebhookUrl].(string); ok {
						t.WebhookUrl = &wu
					}
					_, err := t.PostMessage(ctx, mc, opts)
					if err != nil {
						notificationErrors = append(notificationErrors, err)
					} else {
						externalNotificationSent = true
					}
				}
			}
		}
	}
	return externalNotificationSent, notificationErrors
}

func (ip *Input) consoleIntegration(ctx context.Context, input modconfig.Input, mc MessageCreator, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	c := NewInputIntegrationConsole(NewInputIntegrationBase(ip))
	return c.PostMessage(ctx, mc, options)
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	var inputType, prompt string
	if it, ok := input[schema.AttributeTypeType].(string); ok {
		inputType = it
	}

	if p, ok := input[schema.AttributeTypePrompt].(string); ok {
		prompt = p
	}

	stepName := strings.Split(ip.StepName, ".")[len(strings.Split(ip.StepName, "."))-1]

	return ip.execute(ctx, input, &InputStepMessageCreator{
		Prompt:    prompt,
		InputType: inputType,
		StepName:  stepName,
	})
}

type MessageCreator interface {
	EmailMessage(*InputIntegrationEmail, []InputIntegrationResponseOption) (string, error)
	SlackMessage(*InputIntegrationSlack, []InputIntegrationResponseOption) (slack.Blocks, error)
	MsTeamsMessage(*InputIntegrationMsTeams, []InputIntegrationResponseOption) (*messagecard.MessageCard, error)
	ConsoleMessage(*InputIntegrationConsole, []InputIntegrationResponseOption) (*string, *huh.Form, any, error)
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
		var selectedOptions []*slack.OptionBlockObject
		blockOptions := make([]*slack.OptionBlockObject, len(options))
		for i, opt := range options {
			blockOptions[i] = slack.NewOptionBlockObject(*opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false), nil)
			if opt.Selected != nil && *opt.Selected {
				selectedOptions = append(selectedOptions, blockOptions[i])
			}
		}
		ms := slack.NewOptionsMultiSelectBlockElement(
			slack.MultiOptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, "Select options", false, false), "not_finished", blockOptions...)
		if len(selectedOptions) > 0 {
			ms.InitialOptions = selectedOptions
		}
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

	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	subject := iim.Subject

	if subject == "" {
		subject = icm.Prompt
		if len(subject) > 50 {
			subject = subject[:50] + "..."
		}
	}

	header["Subject"] = subject

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

func (icm *InputStepMessageCreator) MsTeamsMessage(ip *InputIntegrationMsTeams, options []InputIntegrationResponseOption) (*messagecard.MessageCard, error) {
	msgCard := messagecard.NewMessageCard()

	// get response url from cached integration
	i, err := db.GetIntegration(ip.IntegrationName)
	if err != nil {
		return nil, err
	}
	responseUrl := kitTypes.SafeString(i.GetIntegrationImpl().Url)
	if responseUrl == "" {
		return nil, perr.InternalWithMessage("response url not found for integration " + ip.IntegrationName)
	}

	msgCard.Title = icm.Prompt
	msgCard.Summary = icm.Prompt

	pa, err := messagecard.NewPotentialAction(messagecard.PotentialActionActionCardType, ip.StepExecutionID)
	if err != nil {
		return nil, err
	}

	body, err := ip.buildReturnPayload("{{options.value}}", icm.Prompt)
	if err != nil {
		return nil, perr.InternalWithMessage("error building return body payload for " + ip.IntegrationName)
	}

	switch {
	case icm.InputType == constants.InputTypeButton && len(options) <= 4:
		for _, option := range options {
			pa.Actions = append(pa.Actions, messagecard.PotentialActionActionCardAction{
				Type: messagecard.PotentialActionHTTPPostType,
				Name: *option.Label,
				PotentialActionHTTPPOST: messagecard.PotentialActionHTTPPOST{
					Body:    strings.Replace(body, "{{options.value}}", *option.Value, 1),
					Target:  responseUrl,
					Headers: []messagecard.PotentialActionHTTPPOSTHeader{},
				},
			})
		}
	case icm.InputType == constants.InputTypeText:
		pa.Inputs = append(pa.Inputs, messagecard.PotentialActionActionCardInput{
			ID:         "options",
			Type:       messagecard.PotentialActionActionCardInputTextInputType,
			IsRequired: true,
			PotentialActionActionCardInputTextInput: messagecard.PotentialActionActionCardInputTextInput{
				IsMultiline: false,
			}})
		pa.Actions = append(pa.Actions, messagecard.PotentialActionActionCardAction{
			Type: messagecard.PotentialActionHTTPPostType,
			Name: "Submit",
			PotentialActionHTTPPOST: messagecard.PotentialActionHTTPPOST{
				Body:    body,
				Target:  responseUrl,
				Headers: []messagecard.PotentialActionHTTPPOSTHeader{},
			},
		})
	case icm.InputType == constants.InputTypeSelect,
		icm.InputType == constants.InputTypeMultiSelect,
		icm.InputType == constants.InputTypeButton && len(options) > 4:
		isMulti := icm.InputType == constants.InputTypeMultiSelect
		var choices []struct {
			Display string `json:"display,omitempty" yaml:"display,omitempty"`
			Value   string `json:"value,omitempty" yaml:"value,omitempty"`
		}
		for _, option := range options {
			choices = append(choices, struct {
				Display string `json:"display,omitempty" yaml:"display,omitempty"`
				Value   string `json:"value,omitempty" yaml:"value,omitempty"`
			}{
				Display: *option.Label,
				Value:   *option.Value,
			})
		}
		pa.Inputs = append(pa.Inputs, messagecard.PotentialActionActionCardInput{
			ID:         "options",
			Type:       messagecard.PotentialActionActionCardInputMultichoiceInputType,
			IsRequired: true,
			PotentialActionActionCardInputMultichoiceInput: messagecard.PotentialActionActionCardInputMultichoiceInput{
				IsMultiSelect: isMulti,
				Choices:       choices,
				Style:         "expanded",
			}})
		pa.Actions = append(pa.Actions, messagecard.PotentialActionActionCardAction{
			Type: messagecard.PotentialActionHTTPPostType,
			Name: "Submit",
			PotentialActionHTTPPOST: messagecard.PotentialActionHTTPPOST{
				Body:    body,
				Target:  responseUrl,
				Headers: []messagecard.PotentialActionHTTPPOSTHeader{},
			},
		})
	}

	msgCard.PotentialActions = append(msgCard.PotentialActions, pa)
	return msgCard, nil
}

func (icm *InputStepMessageCreator) ConsoleMessage(ip *InputIntegrationConsole, options []InputIntegrationResponseOption) (*string, *huh.Form, any, error) {
	var responseValue any
	var group *huh.Group
	var opts []huh.Option[string]
	for _, opt := range options {
		opts = append(opts, huh.NewOption(*opt.Label, *opt.Value))
	}

	switch icm.InputType {
	case constants.InputTypeButton, constants.InputTypeSelect:
		responseValue = new(string)
		s := huh.NewSelect[string]().Title(icm.Prompt).Options(opts...).Value(responseValue.(*string))
		group = huh.NewGroup(s)
	case constants.InputTypeMultiSelect:
		responseValue = new([]string)
		s := huh.NewMultiSelect[string]().Title(icm.Prompt).Options(opts...).Value(responseValue.(*[]string))
		group = huh.NewGroup(s)
	case constants.InputTypeText:
		responseValue = new(string)
		s := huh.NewInput().Title(icm.Prompt).Value(responseValue.(*string))
		group = huh.NewGroup(s)
	}

	form := huh.NewForm(group)
	return nil, form, responseValue, nil
}
