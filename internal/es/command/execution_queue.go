package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionQueueHandler CommandHandler

func (h ExecutionQueueHandler) HandlerName() string {
	return execution.ExecutionQueueCommand.HandlerName()
}

func (h ExecutionQueueHandler) NewCommand() interface{} {
	return &event.ExecutionQueue{}
}

func (h ExecutionQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.ExecutionQueue)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.ExecutionQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.ExecutionQueue")
	}

	if cmd.PipelineQueue == nil && cmd.TriggerQueue == nil {
		slog.Error("pipeline queue or trigger queue is required")
		return perr.BadRequestWithMessage("pipeline queue or trigger queue is required")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.ExecutionQueuedFromExecutionQueue(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
