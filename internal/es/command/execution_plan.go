package command

import (
	"context"
	"log/slog"
	"slices"

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

	ex, err := execution.GetExecution(cmd.Event.ExecutionID)
	if err != nil {
		slog.Error("Error loading execution", "error", err)
		return err
	}

	// Check if all the pipelines are finished
	allFinished := true
	for _, pex := range ex.PipelineExecutions {
		if !slices.Contains(event.EndEvents, pex.Status) {
			allFinished = false
			continue
		}
	}

	if allFinished {
		// any failure?
		for _, pex := range ex.PipelineExecutions {
			if pex.Status == "failed" {
				// raise execution fail
				cmd := event.ExecutionFailedFromExecutionPlan(cmd)
				err = h.EventBus.Publish(ctx, cmd)
				if err != nil {
					slog.Error("Error publishing event", "error", err)
					return nil
				}
				return nil
			}
		}

		// raise execution finish
		cmd := event.ExecutionFinishedFromExecutionPlan(cmd)
		err = h.EventBus.Publish(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return nil
		}
		return nil
	}

	// Right now there's not much to do in execution plan, we still need to start with either a single
	// pipeline or a trigger
	evt := event.ExecutionPlannedFromExecutionPlan(cmd)

	err = h.EventBus.Publish(ctx, evt)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
		return nil
	}

	return nil
}
