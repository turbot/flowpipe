package primitive

import (
	"context"

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
		Data: map[string]interface{}{
			"container_id": "1234",
			"stdout":       "hello world",
			"stderr":       "",
		},
	}

	return &o, nil
}
