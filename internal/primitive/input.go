package primitive

import (
	"context"
	"encoding/json"

	"github.com/turbot/pipe-fittings/modconfig"
)

type InputType string

const (
	InputTypeSlack InputType = "slack"
	InputTypeEmail InputType = "email"
)

type Input struct {
}

type InputIntegration interface {
	PostMessage(modconfig.Input) error
	ReceiveMessage() (*modconfig.Output, error)
}

type InputIntegrationSlack struct {
}

func (*InputIntegrationSlack) PostMessage(modconfig.Input) error {
	return nil
}

func (*InputIntegrationSlack) ReceiveMessage() (*modconfig.Output, error) {
	return nil, nil
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

	// This is where the actual work is done to setup the approval stuff in slack
	inputType := input["type"].(InputType)

	var err error
	switch inputType {
	case InputTypeSlack:
		slack := InputIntegrationSlack{}
		err = slack.PostMessage(input)

	case InputTypeEmail:
		email := InputIntegrationEmail{}
		err = email.PostMessage(input)
	}

	return &modconfig.Output{}, err
}

func (ip *Input) ProcessOutput(ctx context.Context, requestBody []byte) (*modconfig.Output, error) {

	// TODO: error handling

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
