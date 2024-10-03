package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionStartHandler CommandHandler

func (h ExecutionStartHandler) HandlerName() string {
	return execution.ExecutionStartCommand.HandlerName()
}

func (h ExecutionStartHandler) NewCommand() interface{} {
	return &event.ExecutionStart{}
}

func (h ExecutionStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.ExecutionStart)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.ExecutionStart", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.ExecutionStart")
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

	// Create a new ExecutionStarted event
	e := event.ExecutionStartedFromExecutionStart(cmd)

	return h.EventBus.Publish(ctx, e)
}
