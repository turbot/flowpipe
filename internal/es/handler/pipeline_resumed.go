package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
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

	evt, err := event.NewPipelinePlan(event.ForPipelineResumed(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineResumedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, evt)
}
