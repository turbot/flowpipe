package primitive

import (
	"context"
	"github.com/turbot/flowpipe/internal/resources"
	"log/slog"
	"time"

	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input resources.Input) error {
	if input[schema.AttributeTypeDuration] == nil {
		return perr.BadRequestWithMessage("Sleep input must define a duration")
	}

	switch duration := input[schema.AttributeTypeDuration].(type) {
	case string:
		_, err := time.ParseDuration(duration)
		if err != nil {
			return perr.BadRequestWithMessage("invalid sleep duration " + duration)
		}
	case int64:
		if duration < 0 {
			return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeDuration + "' must be a positive whole number")
		}
	case float64:
		if duration < 0 {
			return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeDuration + "' must be a positive whole number")
		}
	default:
		return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeDuration + "' must be a string or a whole number")
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input resources.Input) (*resources.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	var duration time.Duration
	switch durationVal := input[schema.AttributeTypeDuration].(type) {
	case string:
		duration, _ = time.ParseDuration(durationVal)
	case int64:
		duration = time.Duration(durationVal) * time.Millisecond // in milliseconds
	case float64:
		duration = time.Duration(durationVal) * time.Millisecond // in milliseconds
	}

	slog.Debug("Sleeping for", "duration", duration)
	start := time.Now().UTC()
	time.Sleep(duration)
	finish := time.Now().UTC()

	output := &resources.Output{
		Data: map[string]interface{}{},
	}

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	return output, nil
}
