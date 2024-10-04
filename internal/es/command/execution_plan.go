package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type ExecutionPlanHandler CommandHandler

func (h ExecutionPlanHandler) HandlerName() string {
	return execution.ExecutionPlanCommand.HandlerName()
}

func (h ExecutionPlanHandler) NewCommand() interface{} {
	return &event.ExecutionPlan{}
}

func (h ExecutionPlanHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.ExecutionPlan)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.ExecutionPlan", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.ExecutionPlan")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	// Right now there's not much to do in execution plan, we still need to start with either a single
	// pipeline or a trigger

	e := event.ExecutionPlannedFromExecutionPlan(cmd)

	err := h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
