package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionFinishHandler CommandHandler

func (h ExecutionFinishHandler) HandlerName() string {
	return execution.ExecutionFinishCommand.HandlerName()
}

func (h ExecutionFinishHandler) NewCommand() interface{} {
	return &event.ExecutionFinish{}
}

func (h ExecutionFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.ExecutionFinish)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.ExecutionFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.ExecutionFinish")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.ExecutionFinishedFromExecutionFinish(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
