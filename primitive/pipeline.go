package primitive

import (
	"context"
	"errors"
	"fmt"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type RunPipeline struct{}

func (e *RunPipeline) ValidateInput(ctx context.Context, input pipeline.Input) error {

	if input["name"] == nil {
		return errors.New("pipeline input must define a name")
	}

	pipelineName := input["name"].(string)
	if pipelineName == "" {
		return fmt.Errorf("invalid pipeline name: %s", pipelineName)
	}

	if args, ok := input["args"].(map[string]interface{}); !ok {
		return fmt.Errorf("pipeline args must be a map of values to arg name: %s", args)
	}

	return nil
}

func (e *RunPipeline) Run(ctx context.Context, input pipeline.Input) (*pipeline.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := &pipeline.Output{
		"name": input["name"].(string),
		"args": input["args"].(map[string]interface{}),
	}

	return output, nil
}
