package primitive

// TODO: development primitive - remove

import (
	"context"

	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Echo struct{}

func (e *Echo) ValidateInput(ctx context.Context, i types.Input) error {
	return nil
}

func (e *Echo) Run(ctx context.Context, input types.Input) (*types.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	o := types.Output{
		Data: map[string]interface{}{},
	}
	o.Data[schema.AttributeTypeText] = input[schema.AttributeTypeText]
	o.Data[schema.AttributeTypeJson] = input[schema.AttributeTypeJson]

	return &o, nil
}
