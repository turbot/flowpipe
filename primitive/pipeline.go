package primitive

import (
	"context"
	"errors"
	"fmt"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type RunPipeline struct{}

func (e *RunPipeline) ValidateInput(ctx context.Context, input pipeline.StepInput) error {

	if input["name"] == nil {
		return errors.New("pipeline input must define a name")
	}

	pipelineName := input["name"].(string)
	if pipelineName == "" {
		return fmt.Errorf("invalid pipeline name: %s", pipelineName)
	}

	return nil
}

func (e *RunPipeline) Run(ctx context.Context, input pipeline.StepInput) (pipeline.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := pipeline.StepOutput{
		"name": input["name"].(string),
		// TODO - needs to pass the actual input
		"input": pipeline.PipelineInput{},
	}

	return output, nil
}
