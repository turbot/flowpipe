package primitive

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
)

type RunPipeline struct{}

func (e *RunPipeline) ValidateInput(ctx context.Context, input types.Input) error {

	if input["name"] == nil {
		return fperr.BadRequestWithMessage("pipeline input must define a name")
	}

	pipelineName := input["name"].(string)
	if pipelineName == "" {
		return fperr.BadRequestWithMessage("invalid pipeline name: " + pipelineName)
	}

	if _, ok := input["args"].(map[string]interface{}); !ok {
		return fperr.BadRequestWithMessage("pipeline args must be a map of values to arg name")
	}

	return nil
}

func (e *RunPipeline) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := &types.StepOutput{
		OutputVariables: map[string]interface{}{},
	}

	output.OutputVariables["name"] = input["name"].(string)
	output.OutputVariables["args"] = input["args"].(map[string]interface{})

	return output, nil
}
