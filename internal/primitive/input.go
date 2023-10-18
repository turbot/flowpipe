package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/slack-go/slack"
	"github.com/turbot/pipe-fittings/modconfig"
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
	// Slack User Token - Should be from the input config
	api := slack.New("xoxp-2556146250-518904151623-6045632668293-eed618a227d1918f1ee82f72761b28bd")

	// Channel ID - Should be from the input config
	channelID := "DF8SL4GR5"

	payload := map[string]interface{}{
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
		"execution_id":          ip.ExecutionID,
	}
	test, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	encodedText := base64.StdEncoding.EncodeToString(test)

	// Check if the user has already made a selection
	userHasMadeSelection := false

	// Send the input prompt
	attachment := slack.Attachment{
		// Fallback:   "exec_id=12345677",
		Text:       "Choose an option:",
		Color:      "#3AA3E3",
		CallbackID: encodedText,
		Actions: []slack.AttachmentAction{
			{
				Name:  "YES",
				Text:  "Yes",
				Type:  "button",
				Value: "Yes",
			},
			{
				Name:  "NO",
				Text:  "No",
				Type:  "button",
				Value: "No",
			},
		},
	}

	// Remove the interactive element if the user has made a selection
	if userHasMadeSelection {
		attachment.Actions = nil
	}

	_, _, err = api.PostMessage(channelID,
		slack.MsgOptionMetadata(slack.SlackMetadata{
			EventPayload: map[string]interface{}{
				"pipeline_execution_id": ip.PipelineExecutionID,
			},
		}),
		slack.MsgOptionText("Please make a selection:", false),
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

	// test, err := os.ReadFile("./test.json")
	// if err != nil {
	// 	return nil, err
	// }

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
	return nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// return &modconfig.Output{}, nil

	// This is where the actual work is done to setup the approval stuff in slack
	// inputType := input["type"].(InputType)
	// TODO: Remove hardcoding
	inputType := InputTypeSlack

	var err error
	switch inputType {
	case InputTypeSlack:
		slack := InputIntegrationSlack{
			InputIntegrationBase: InputIntegrationBase{
				ExecutionID:         ip.ExecutionID,
				PipelineExecutionID: ip.PipelineExecutionID,
				StepExecutionID:     ip.StepExecutionID,
			},
		}
		err = slack.PostMessage(input)

		// _, err = slack.ReceiveMessage(ctx, nil)

	case InputTypeEmail:
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
