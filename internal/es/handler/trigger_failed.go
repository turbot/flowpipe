package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerFailed EventHandler

func (h TriggerFailed) HandlerName() string {
	return execution.TriggerFailedEvent.HandlerName()
}

func (h TriggerFailed) NewEvent() interface{} {
	return &event.TriggerFailed{}
}
func (h TriggerFailed) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.TriggerFailed)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.TriggerFailed", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.TriggerFailed")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	// There's only 1 trigger for each execution, it's a straight forward process to
	// fail the execution
	cmd := event.ExecutionFailFromTriggerFailed(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
