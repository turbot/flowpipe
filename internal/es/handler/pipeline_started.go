package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PipelineStarted EventHandler

func (h PipelineStarted) HandlerName() string {
	return "handler.pipeline_started"
}

func (PipelineStarted) NewEvent() interface{} {
	return &event.PipelineStarted{}
}

func (h PipelineStarted) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineStarted)

	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.PipelineStarted", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineStarted")
	}

	evt, err := event.NewPipelinePlan(event.ForPipelineStarted(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStartedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, evt)
}
