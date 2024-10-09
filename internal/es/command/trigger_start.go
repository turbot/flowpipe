package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerStartHandler CommandHandler

func (h TriggerStartHandler) HandlerName() string {
	return execution.TriggerStartCommand.HandlerName()
}

func (h TriggerStartHandler) NewCommand() interface{} {
	return &event.TriggerStart{}
}

func (h TriggerStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.TriggerStart)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.TriggerStart", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.TriggerStart")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	trg, err := db.GetTrigger(cmd.Name)
	if err != nil {
		slog.Error("Error getting trigger", "error", err)
		evt := event.TriggerFailedFromTriggerStart(cmd)
		err := h.EventBus.Publish(ctx, evt)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return nil
	}

	evt := event.TriggerStartedFromTriggerStart(cmd, trg)
	err = h.EventBus.Publish(ctx, evt)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}

	return nil
}
