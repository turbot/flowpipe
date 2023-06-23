package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return "handler.pipeline_queued"
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
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineQueued")
	}

	logger.Info("[10] pipeline_queued event handler", "executionID", e.Event.ExecutionID)

	cmd, err := event.NewPipelineLoad(event.ForPipelineQueued(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
