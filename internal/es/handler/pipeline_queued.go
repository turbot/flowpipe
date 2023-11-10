package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return execution.PipelineQueuedEvent.HandlerName()
}

func (PipelineQueued) NewEvent() interface{} {
	return &event.PipelineQueued{}
}

// Path from here:
// * PipelineQueued -> PipelineLoad command -> PipelineLoaded event handler
//
// ? is this meant to be when something is being picked up from the queue?
// ? so the PipelineQueue *command* is the one that puts it in a some sort of a queue?
func (h PipelineQueued) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineQueued)

	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineQueued")
	}

	cmd, err := event.NewPipelineLoad(event.ForPipelineQueued(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
