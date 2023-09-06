package primitive

import (
	"context"

	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type RunPipeline struct{}

func (e *RunPipeline) ValidateInput(ctx context.Context, input modconfig.Input) error {

	if input[schema.AttributeTypePipeline] == nil {
		return perr.BadRequestWithMessage("pipeline input must define a name")
	}

	pipelineName := input[schema.AttributeTypePipeline].(string)
	if pipelineName == "" {
		return perr.BadRequestWithMessage("invalid pipeline name: " + pipelineName)
	}

	if input[schema.AttributeTypeArgs] != nil {
		if _, ok := input[schema.AttributeTypeArgs].(map[string]interface{}); !ok {
			return perr.BadRequestWithMessage("pipeline args must be a map of values to arg name")
		}
	}

	return nil
}

func (e *RunPipeline) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	output := &modconfig.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypePipeline] = input[schema.AttributeTypePipeline].(string)

	if input[schema.AttributeTypeArgs] != nil {
		output.Data[schema.AttributeTypeArgs] = input[schema.AttributeTypeArgs].(map[string]interface{})
	}

	return output, nil
}
