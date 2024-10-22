package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionFailHandler CommandHandler

func (h ExecutionFailHandler) HandlerName() string {
	return execution.ExecutionFailCommand.HandlerName()
}

func (h ExecutionFailHandler) NewCommand() interface{} {
	return &event.ExecutionFail{}
}

func (h ExecutionFailHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.ExecutionFail)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.ExecutionFail", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.ExecutionFail")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.ExecutionFailedFromExecutionFail(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
