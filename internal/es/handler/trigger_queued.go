package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerQueued EventHandler

func (h TriggerQueued) HandlerName() string {
	return execution.TriggerQueuedEvent.HandlerName()
}

func (h TriggerQueued) NewEvent() interface{} {
	return &event.TriggerQueued{}
}

func (h TriggerQueued) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.TriggerQueued)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.TriggerQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.TriggerQueued")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	cmd := event.TriggerStartFromTriggerQueued(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
