package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerFinishHandler CommandHandler

func (h TriggerFinishHandler) HandlerName() string {
	return execution.TriggerFinishCommand.HandlerName()
}

func (h TriggerFinishHandler) NewCommand() interface{} {
	return &event.TriggerFinish{}
}

func (h TriggerFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.TriggerFinish)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.TriggerFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.TriggerFinish")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.TriggerFinishedFromTriggerFinish(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
