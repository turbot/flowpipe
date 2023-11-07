package primitive

// TODO: development primitive - remove

import (
	"context"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

type Transform struct{}

func (e *Transform) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (e *Transform) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{},
	}
	o.Data[schema.AttributeTypeValue] = input[schema.AttributeTypeValue]

	return &o, nil
}
