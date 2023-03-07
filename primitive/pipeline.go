package primitive

import (
	"context"
	"errors"
	"fmt"
)

type RunPipeline struct{}

func (e *RunPipeline) ValidateInput(ctx context.Context, input Input) error {

	if input["name"] == nil {
		return errors.New("pipeline input must define a name")
	}

	pipelineName := input["name"].(string)
	if pipelineName == "" {
		return fmt.Errorf("invalid pipeline name: %s", pipelineName)
	}

	return nil
}

func (e *RunPipeline) Run(ctx context.Context, input Input) (Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := Output{
		"name": input["name"].(string),
	}

	return output, nil
}
