package primitive

// TODO: development primitive - remove

import (
	"context"

	"github.com/turbot/flowpipe/pipeparser/pipeline"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Echo struct{}

func (e *Echo) ValidateInput(ctx context.Context, i pipeline.Input) error {
	return nil
}

func (e *Echo) Run(ctx context.Context, input pipeline.Input) (*pipeline.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	o := pipeline.Output{
		Data: map[string]interface{}{},
	}
	o.Data[schema.AttributeTypeText] = input[schema.AttributeTypeText]
	o.Data[schema.AttributeTypeJson] = input[schema.AttributeTypeJson]

	return &o, nil
}
