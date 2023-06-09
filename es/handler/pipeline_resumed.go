package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineResumed EventHandler

func (h PipelineResumed) HandlerName() string {
	return "handler.pipeline_resumed"
}

func (PipelineResumed) NewEvent() interface{} {
	return &event.PipelineResumed{}
}

func (h PipelineResumed) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelineResumed)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineResumed", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineResumed")
	}

	logger.Info("[11] pipeline_resumed event handler", "eventExecutionID", e.Event.ExecutionID)

	evt, err := event.NewPipelinePlan(event.ForPipelineResumed(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineResumedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, evt)
}
