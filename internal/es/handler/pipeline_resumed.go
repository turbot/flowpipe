package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineResumed EventHandler

var pipelineResumed = event.PipelineResumed{}

func (h PipelineResumed) HandlerName() string {
	return pipelineResumed.HandlerName()
}

func (PipelineResumed) NewEvent() interface{} {
	return &event.PipelineResumed{}
}

func (h PipelineResumed) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelineResumed)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineResumed", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineResumed")
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineResumed(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineResumedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
