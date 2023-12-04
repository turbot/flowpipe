package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
	"log/slog"
)

type PipelineResumed EventHandler

func (h PipelineResumed) HandlerName() string {
	return execution.PipelineResumedEvent.HandlerName()
}

func (PipelineResumed) NewEvent() interface{} {
	return &event.PipelineResumed{}
}

func (h PipelineResumed) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineResumed)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineResumed", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineResumed")
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineResumed(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineResumedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
