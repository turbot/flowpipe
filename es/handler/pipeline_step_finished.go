package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineStepFinished EventHandler

func (h PipelineStepFinished) HandlerName() string {
	return "handler.pipeline_step_finished"
}

func (PipelineStepFinished) NewEvent() interface{} {
	return &event.PipelineStepFinished{}
}

func (h PipelineStepFinished) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineStepFinished)
	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.PipelineStepFinished", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineStepFinished")
	}

	fplog.Logger(ctx).Info("[13] pipeline_step_finished event handler", "executionID", e.Event.ExecutionID)
	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
