package primitive

import (
	"context"
	"encoding/json"

	"github.com/turbot/flowpipe/pipeparser/modconfig"
)

type Input struct{}

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
