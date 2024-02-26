package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

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
	boldPromptBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*%s*", prompt), false, false)

	switch inputType {
	case constants.InputTypeButton:
		header := slack.NewSectionBlock(boldPromptBlock, nil, nil, slack.SectionBlockOptionBlockID(encodedPayload))
		var buttons []slack.BlockElement
		for i, opt := range options {
			button := slack.NewButtonBlockElement(fmt.Sprintf("finished_%d", i), *opt.Value, slack.NewTextBlockObject(slack.PlainTextType, *opt.Label, false, false))
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
		return nil, perr.InternalWithMessage(fmt.Sprintf("Type %s not yet implemented for Slack Integration", inputType))
	}

	output := modconfig.Output{}
	if !helpers.IsNil(ip.Token) && !helpers.IsNil(ip.Channel) {
		msgOption := slack.MsgOptionBlocks(blocks.BlockSet...)
		api := slack.New(*ip.Token)
		_, _, err = api.PostMessage(*ip.Channel, msgOption, slack.MsgOptionAsUser(false))
		return &output, err
	} else {
		wMsg := slack.WebhookMessage{Blocks: &blocks}
		err = slack.PostWebhook(*ip.WebhookUrl, &wMsg)
		return &output, err
	}
}
