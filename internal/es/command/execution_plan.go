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

	// Check if we have started the trigger execution
	if cmd.TriggerQueue != nil && ex.TriggerExecution == nil && len(ex.PipelineExecutions) == 0 {

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

	// Check if this is the start of pipeline queue
	if cmd.PipelineQueue != nil && len(ex.PipelineExecutions) == 0 {
		// Pipeline hasn't started yet
		evt := event.ExecutionPlannedFromExecutionPlan(cmd)

		err = h.EventBus.Publish(ctx, evt)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return nil
		}

		return nil
	}

	// check if all pipelines are paused
	allPaused := true
	for _, pex := range ex.PipelineExecutions {
		if pex.Status != "paused" {
			allPaused = false
			break
		}
	}

	if allPaused {
		// raise execution paused
		cmd := event.ExecutionPausedFromExecutionPlan(cmd)
		err = h.EventBus.Publish(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return nil
		}
		return nil
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

		failure := false

		// any failure?
		for _, pex := range ex.PipelineExecutions {
			if pex.Status == "failed" {
				failure = true
				break
			}
		}

		if ex.TriggerExecution != nil {
			if failure {
				// raise trigger fail
				cmd := event.TriggerFailedFromExecutionPlan(cmd, ex.TriggerExecution.Name)
				err = h.EventBus.Publish(ctx, cmd)
				if err != nil {
					slog.Error("Error publishing event", "error", err)
					return nil
				}
				return nil
			} else {
				// raise trigger finish
				cmd := event.TriggerFinishedFromExecutionPlan(cmd, ex.TriggerExecution.Name)
				err = h.EventBus.Publish(ctx, cmd)
				if err != nil {
					slog.Error("Error publishing event", "error", err)
					return nil
				}
				return nil
			}
		} else {
			if failure {
				// raise execution fail
				cmd := event.ExecutionFailedFromExecutionPlan(cmd, perr.InternalWithMessage("pipeline failed"))
				err = h.EventBus.Publish(ctx, cmd)
				if err != nil {
					slog.Error("Error publishing event", "error", err)
					return nil
				}
				return nil
			} else {
				// raise execution finish
				cmd := event.ExecutionFinishedFromExecutionPlan(cmd)
				err = h.EventBus.Publish(ctx, cmd)
				if err != nil {
					slog.Error("Error publishing event", "error", err)
					return nil
				}
				return nil
			}
		}

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
