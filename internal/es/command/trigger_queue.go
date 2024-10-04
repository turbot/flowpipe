package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerQueueHandler CommandHandler

func (h TriggerQueueHandler) HandlerName() string {
	return execution.TriggerQueueCommand.HandlerName()
}

func (h TriggerQueueHandler) NewCommand() interface{} {
	return &event.TriggerQueue{}
}

func (h TriggerQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.TriggerQueue)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.TriggerQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.TriggerQueue")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.TriggerQueuedFromTriggerQueue(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
