package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionPlanned EventHandler

func (h ExecutionPlanned) HandlerName() string {
	return execution.ExecutionPlannedEvent.HandlerName()
}

func (h ExecutionPlanned) NewEvent() interface{} {
	return &event.ExecutionPlanned{}
}

func (h ExecutionPlanned) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.ExecutionPlanned)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionPlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionPlanned")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	if evt.TriggerQueue != nil {
		if evt.TriggerQueue.Event == nil {
			evt.TriggerQueue.Event = evt.Event
		}

		err := h.CommandBus.Send(ctx, evt.TriggerQueue)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return nil
		}

		return nil
	} else if evt.PipelineQueue != nil {
		err := h.CommandBus.Send(ctx, evt.PipelineQueue)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return nil
		}

		return nil
	}

	cmd := event.ExecutionFinishFromExecutionPlanned(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}

	return nil
}
