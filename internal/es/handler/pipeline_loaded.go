package handler

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineLoaded EventHandler

func (h PipelineLoaded) HandlerName() string {
	return execution.PipelineLoadedEvent.HandlerName()
}

func (PipelineLoaded) NewEvent() interface{} {
	return &event.PipelineLoaded{}
}

func (h PipelineLoaded) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.PipelineLoaded)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineLoaded")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	cmd, err := event.NewPipelineStart(event.ForPipelineLoaded(evt))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineLoadedToPipelineFail(evt, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
