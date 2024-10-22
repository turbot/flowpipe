package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionFinished EventHandler

func (h ExecutionFinished) HandlerName() string {
	return execution.ExecutionFinishedEvent.HandlerName()
}

func (h ExecutionFinished) NewEvent() interface{} {
	return &event.ExecutionFinished{}
}

func (h ExecutionFinished) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.ExecutionFinished)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.ExecutionFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.ExecutionFinished")
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
