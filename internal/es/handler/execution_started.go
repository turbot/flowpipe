package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionStarted EventHandler

func (h ExecutionStarted) HandlerName() string {
	return execution.ExecutionStartedEvent.HandlerName()
}

func (h ExecutionStarted) NewEvent() interface{} {
	return &event.ExecutionStarted{}
}

func (h ExecutionStarted) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.ExecutionStarted)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionStarted")
	}
	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	cmd := event.ExecutionPlanFromExecutionStarted(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
