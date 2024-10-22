package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionPaused EventHandler

func (h ExecutionPaused) HandlerName() string {
	return execution.ExecutionPausedEvent.HandlerName()
}

func (h ExecutionPaused) NewEvent() interface{} {
	return &event.ExecutionPaused{}
}

func (h ExecutionPaused) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.ExecutionPaused)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionPaused", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionPaused")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	return nil
}
