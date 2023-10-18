package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/slack-go/slack"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type InputType string

const (
	InputTypeSlack InputType = "slack"
	InputTypeEmail InputType = "email"
)

type Input struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

type InputIntegration interface {
	PostMessage(modconfig.Input) error
	ReceiveMessage() (*modconfig.Output, error)
}

type InputIntegrationSlack struct {
	InputIntegrationBase
}

type InputIntegrationBase struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

func (ip *InputIntegrationSlack) PostMessage(input modconfig.Input) error {
	// Set the slack user token provided in the input config
	userToken := input[schema.AttributeTypeToken].(string)
	api := slack.New(userToken)

	// Set the slack user token provided in the input config
	channelID := input[schema.AttributeTypeChannel].(string)

	// Encode the callback_id to pass to the interactive element
	payload := map[string]interface{}{
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
		"execution_id":          ip.ExecutionID,
	}
	unmarshaledAdditionalInfo, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	encodedText := base64.StdEncoding.EncodeToString(unmarshaledAdditionalInfo)

	// Check if the user has already made a selection
	userHasMadeSelection := false

	var attachment slack.Attachment

	// TODO: add validation for the slack type and prompt
	slackType := input[schema.AttributeTypeSlackType].(string)
	prompt := input[schema.AttributeTypePrompt].(string)

	options := []string{"Approve", "Reject", "Ignore"}

	// Check for the prompt
	if slackType == "button" {
		attachment = slack.Attachment{
			Text:       prompt,
			Color:      "#3AA3E3",
			CallbackID: encodedText,
			// Actions: []slack.AttachmentAction{
			// 	{
			// 		Name:  "Approve",
			// 		Text:  "Approve",
			// 		Type:  "button",
			// 		Value: "Approve",
			// 	},
			// 	{
			// 		Name:  "Reject",
			// 		Text:  "Reject",
			// 		Type:  "button",
			// 		Value: "Reject",
			// 	},
			// },
		}

		var actions []slack.AttachmentAction
		for _, opt := range options {
			actions = append(actions, slack.AttachmentAction{
				Name:  opt,
				Text:  opt,
				Type:  "button",
				Value: opt,
			})
		}

		if len(actions) > 0 {
			attachment.Actions = actions
		}
	}

	// Remove the interactive element if the user has made a selection
	if userHasMadeSelection {
		attachment.Actions = nil
	}

	_, _, err = api.PostMessage(channelID,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true))

	if err != nil {
		return err
	}
	return nil
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

func (*InputIntegrationSlack) ReceiveMessage(ctx context.Context, requestBody []byte) (*modconfig.Output, error) {
	var bodyJSON map[string]interface{}

	err := json.Unmarshal(requestBody, &bodyJSON)
	if err != nil {
		return nil, err
	}

	// Decode the callback_id to extract the execution_id, pipeline_execution_id and step_execution_id
	rawDecodedText, err := base64.StdEncoding.DecodeString(bodyJSON["callback_id"].(string))
	if err != nil {
		return nil, err
	}

	var decodedText JSONPayload
	err = json.Unmarshal(rawDecodedText, &decodedText)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"pipeline_execution_id": decodedText.PipelineExecutionID,
		"step_execution_id":     decodedText.StepExecutionID,
		"execution_id":          decodedText.ExecutionID,
	}

	// If the input type isa button, then the value is a string
	// if the input is multi-select box, then the value is a list
	var value interface{}
	if bodyJSON["actions"] != nil {
		for _, action := range bodyJSON["actions"].([]interface{}) {
			if action.(map[string]interface{})["type"] == "button" {
				value = action.(map[string]interface{})["value"]
			}
		}
	}
	data["value"] = value

	o := modconfig.Output{
		Data: data,
	}

	return &o, nil
}

type InputIntegrationEmail struct {
}

func (*InputIntegrationEmail) PostMessage(modconfig.Input) error {
	return nil
}

func (*InputIntegrationEmail) ReceiveMessage() (*modconfig.Output, error) {
	return nil, nil
}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {

	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Slack input must define a type")
	}

	if _, ok := i[schema.AttributeTypeType].(string); !ok {
		return perr.BadRequestWithMessage("Slack input type must be a string")
	}
	inputType := i[schema.AttributeTypeType].(string)

	switch inputType {
	case string(InputTypeSlack):
		// Validate token
		if i[schema.AttributeTypeToken] == nil {
			return perr.BadRequestWithMessage("Slack input must define a token")
		}
		if _, ok := i[schema.AttributeTypeToken].(string); !ok {
			return perr.BadRequestWithMessage("Slack input token must be a string")
		}

		// Validate channel
		if i[schema.AttributeTypeChannel] == nil {
			return perr.BadRequestWithMessage("Slack input must define a channel")
		}
		if _, ok := i[schema.AttributeTypeChannel].(string); !ok {
			return perr.BadRequestWithMessage("Slack input channel must be a string")
		}

		// Validate the prompt
		if i[schema.AttributeTypePrompt] == nil {
			return perr.BadRequestWithMessage("Slack input must define a prompt")
		}
		if _, ok := i[schema.AttributeTypePrompt].(string); !ok {
			return perr.BadRequestWithMessage("Slack input prompt must be a string")
		}

		// Validate the slack type
		if i[schema.AttributeTypeSlackType] == nil {
			return perr.BadRequestWithMessage("Slack input must define a slack type")
		}
		if _, ok := i[schema.AttributeTypeSlackType].(string); !ok {
			return perr.BadRequestWithMessage("Slack input slack type must be a string")
		}
	case string(InputTypeEmail):
	}

	return nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// This is where the actual work is done to setup the approval stuff in slack
	inputType := input["type"].(string)

	var err error
	switch inputType {
	case string(InputTypeSlack):
		slack := InputIntegrationSlack{
			InputIntegrationBase: InputIntegrationBase{
				ExecutionID:         ip.ExecutionID,
				PipelineExecutionID: ip.PipelineExecutionID,
				StepExecutionID:     ip.StepExecutionID,
			},
		}
		err = slack.PostMessage(input)

	case string(InputTypeEmail):
		email := InputIntegrationEmail{}
		err = email.PostMessage(input)
	}

	return &modconfig.Output{}, err
}

func (ip *Input) ProcessOutput(ctx context.Context, inputType InputType, requestBody []byte) (*modconfig.Output, error) {

	// TODO: error handling
	switch inputType {
	case InputTypeSlack:
		slack := InputIntegrationSlack{}
		_, err := slack.ReceiveMessage(ctx, requestBody)
		if err != nil {
			return nil, err
		}
	case InputTypeEmail:
		// case Input
	}

	var bodyJSON map[string]interface{}
	err := json.Unmarshal(requestBody, &bodyJSON)
	if err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: bodyJSON,
	}

	return &o, nil
}
