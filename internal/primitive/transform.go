package primitive

// TODO: development primitive - remove

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/schema"
)

type Transform struct{}

func (e *Transform) ValidateInput(ctx context.Context, i resources.Input) error {
	return nil
}

func (e *Transform) Run(ctx context.Context, input resources.Input) (*resources.Output, error) {
	start := time.Now().UTC()

	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := resources.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeValue] = input[schema.AttributeTypeValue]
	finish := time.Now().UTC()

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	return &output, nil
}
