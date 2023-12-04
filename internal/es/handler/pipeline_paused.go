package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
	"log/slog"
)

type PipelinePaused EventHandler

func (h PipelinePaused) HandlerName() string {
	return execution.PipelinePausedEvent.HandlerName()
}

func (PipelinePaused) NewEvent() interface{} {
	return &event.PipelinePaused{}
}

func (h PipelinePaused) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelinePaused)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelinePaused", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePaused")
	}

	event.ReleaseEventLogMutex(e.Event.ExecutionID)
	return nil
}
