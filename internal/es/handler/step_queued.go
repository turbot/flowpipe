package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueued EventHandler

func (h StepQueued) HandlerName() string {
	return execution.StepQueuedEvent.HandlerName()
}

func (StepQueued) NewEvent() interface{} {
	return &event.StepQueued{}
}

func (h StepQueued) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.StepQueued)

	if !ok {
		logger.Error("invalid event type", "expected", "*event.StepQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepQueued")
	}

	// Step has been queued (but not yet started), so here we just need to start the step
	// the code should be the same as the pipeline_planned event handler
	cmd, err := event.NewStepStartFromStepQueued(e)

	if err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepQueuedToPipelineFail(e, err)))
		if err != nil {
			fplog.Logger(ctx).Error("Error publishing event", "error", err)
		}

		return nil
	}

	if err := h.CommandBus.Send(ctx, cmd); err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepQueuedToPipelineFail(e, err)))
		if err != nil {
			fplog.Logger(ctx).Error("Error publishing event", "error", err)
		}
		return nil
	}

	return nil
}
