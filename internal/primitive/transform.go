package primitive

// TODO: development primitive - remove

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"time"

	"github.com/turbot/pipe-fittings/schema"
)

type Transform struct{}

func (e *Transform) ValidateInput(ctx context.Context, i flowpipe.Input) error {
	return nil
}

func (e *Transform) Run(ctx context.Context, input flowpipe.Input) (*flowpipe.Output, error) {
	start := time.Now().UTC()

	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := flowpipe.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeValue] = input[schema.AttributeTypeValue]
	finish := time.Now().UTC()

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	return &output, nil
}
