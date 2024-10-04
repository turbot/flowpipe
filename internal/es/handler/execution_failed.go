package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionFailed EventHandler

func (h ExecutionFailed) HandlerName() string {
	return execution.ExecutionFailedEvent.HandlerName()
}

func (h ExecutionFailed) NewEvent() interface{} {
	return &event.ExecutionFailed{}
}

func (h ExecutionFailed) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.ExecutionFailed)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionFailed", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionFailed")
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
