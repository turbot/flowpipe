package handler

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineResumed EventHandler

func (h PipelineResumed) HandlerName() string {
	return execution.PipelineResumedEvent.HandlerName()
}

func (PipelineResumed) NewEvent() interface{} {
	return &event.PipelineResumed{}
}

func (h PipelineResumed) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.PipelineResumed)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineResumed", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineResumed")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	cmd, err := event.NewPipelinePlan(event.ForPipelineResumed(evt))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineResumedToPipelineFail(evt, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
