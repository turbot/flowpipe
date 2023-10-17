package primitive

import (
	"context"
	"encoding/json"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type InputType string

const (
	InputTypeSlack InputType = "slack"
	InputTypeEmail InputType = "email"
)

type Input struct {
	InputType        InputType
	InputIntegration InputIntegration
}

func NewInputPrimitive(inputType InputType) (*Input, error) {
	switch inputType {
	case InputTypeSlack:
		return &Input{
			InputType:        InputTypeSlack,
			InputIntegration: &InputIntegrationSlack{},
		}, nil
	case InputTypeEmail:
		return &Input{
			InputType:        InputTypeEmail,
			InputIntegration: &InputIntegrationEmail{},
		}, nil

	default:
		return nil, perr.BadRequestWithMessage("invalid input type: " + string(inputType))
	}

}

type InputIntegration interface {
	PostMessage() error
	ReceiveMessage() (*modconfig.Output, error)
}

type InputIntegrationSlack struct {
}

func (*InputIntegrationSlack) PostMessage() error {
	return nil
}

func (*InputIntegrationSlack) ReceiveMessage() (*modconfig.Output, error) {
	return nil, nil
}

type InputIntegrationEmail struct {
}

func (*InputIntegrationEmail) PostMessage() error {
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

	o := modconfig.Output{
		Data: map[string]interface{}{},
	}

	return &o, nil
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
