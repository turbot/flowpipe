package primitive

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/configschema"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input types.Input) error {

	if input[configschema.AttributeTypeDuration] == nil {
		return fperr.BadRequestWithMessage("Sleep input must define a duration")
	}

	// json umarshalling converts numbers to float64
	_, ok := input[configschema.AttributeTypeDuration].(float64)

	if !ok {
		fplog.Logger(ctx).Error("invalid sleep duration", "duration", input[configschema.AttributeTypeDuration])
		return fperr.BadRequestWithMessage("invalid sleep duration")
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	durationSecond := input["duration"].(float64)
	// Already validated
	duration := time.Duration(durationSecond * float64(time.Second))

	fplog.Logger(ctx).Info("Sleeping for", "duration", duration)
	start := time.Now().UTC()
	time.Sleep(duration)
	finish := time.Now().UTC()

	// TODO: remove the "duration" output, it's not needed, I added to help with debugging
	return &types.StepOutput{"started_at": start, "finished_at": finish, "duration": duration}, nil
}
