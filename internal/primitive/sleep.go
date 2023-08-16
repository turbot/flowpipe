package primitive

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input pipeline.Input) error {

	if input[schema.AttributeTypeDuration] == nil {
		return pcerr.BadRequestWithMessage("Sleep input must define a duration")
	}

	durationString := input[schema.AttributeTypeDuration].(string)
	_, err := time.ParseDuration(durationString)
	if err != nil {
		return pcerr.BadRequestWithMessage("invalid sleep duration " + durationString)
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input pipeline.Input) (*pipeline.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	durationString := input[schema.AttributeTypeDuration].(string)
	// Already validated
	duration, _ := time.ParseDuration(durationString)

	fplog.Logger(ctx).Info("Sleeping for", "duration", duration)
	start := time.Now().UTC()
	time.Sleep(duration)
	finish := time.Now().UTC()

	output := &pipeline.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeStartedAt] = start
	output.Data[schema.AttributeTypeFinishedAt] = finish
	output.Data[schema.AttributeTypeDuration] = durationString

	return output, nil
}
