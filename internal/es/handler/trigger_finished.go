package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerFinished EventHandler

func (h TriggerFinished) HandlerName() string {
	return execution.TriggerFinishedEvent.HandlerName()
}

func (h TriggerFinished) NewEvent() interface{} {
	return &event.TriggerFinished{}
}

func (h TriggerFinished) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.TriggerFinished)

	if !ok {
		return perr.BadRequestWithMessage("invalid event type expected *event.TriggerFinished")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	cmd := event.ExecutionFinishFromTriggerFinished(evt)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		return nil
	}

	return nil
}
