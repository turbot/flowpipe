package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionQueued EventHandler

func (h ExecutionQueued) HandlerName() string {
	return execution.ExecutionQueuedEvent.HandlerName()
}

func (h ExecutionQueued) NewEvent() interface{} {
	return &event.ExecutionQueued{}
}

func (h ExecutionQueued) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.ExecutionQueued)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionQueued")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	cmd := event.ExecutionStartFromExecutionQueued(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
