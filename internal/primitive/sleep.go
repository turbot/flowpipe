package primitive

import (
	"context"
	"log/slog"
	"time"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input modconfig.Input) error {

	if input[schema.AttributeTypeDuration] == nil {
		return perr.BadRequestWithMessage("Sleep input must define a duration")
	}

	durationString := input[schema.AttributeTypeDuration].(string)
	_, err := time.ParseDuration(durationString)
	if err != nil {
		return perr.BadRequestWithMessage("invalid sleep duration " + durationString)
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	durationString := input[schema.AttributeTypeDuration].(string)
	// Already validated
	duration, _ := time.ParseDuration(durationString)

	slog.Info("Sleeping for", "duration", duration)
	start := time.Now().UTC()
	time.Sleep(duration)
	finish := time.Now().UTC()

	output := &modconfig.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeStartedAt] = start
	output.Data[schema.AttributeTypeFinishedAt] = finish
	output.Data[schema.AttributeTypeDuration] = durationString

	return output, nil
}
