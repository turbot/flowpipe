package primitive

// TODO: development primitive - remove

import (
	"context"

	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Echo struct{}

func (e *Echo) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (e *Echo) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{},
	}
	o.Data[schema.AttributeTypeText] = input[schema.AttributeTypeText]
	o.Data[schema.AttributeTypeJson] = input[schema.AttributeTypeJson]
	o.Data[schema.AttributeTypeNumeric] = input[schema.AttributeTypeNumeric]

	return &o, nil
}
