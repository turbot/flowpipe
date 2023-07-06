package primitive

// TODO: development primitive - remove

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
)

type Text struct{}

func (e *Text) ValidateInput(ctx context.Context, i types.Input) error {
	if i["text"] == nil {
		return fperr.BadRequestWithMessage("Text input must define text")
	}
	return nil
}

func (e *Text) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	return &types.StepOutput{
		"result": input["text"],
	}, nil
}
