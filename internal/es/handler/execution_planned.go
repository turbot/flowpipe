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

	return nil
}
