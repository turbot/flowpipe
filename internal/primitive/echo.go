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

func (e *Echo) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	return &types.StepOutput{
		schema.AttributeTypeText: input[schema.AttributeTypeText],
	}, nil
}
