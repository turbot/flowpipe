package command

import (
	"context"
	"log/slog"
	"time"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueueHandler CommandHandler

func (h StepQueueHandler) HandlerName() string {
	return execution.StepQueueCommand.HandlerName()
}

func (h StepQueueHandler) NewCommand() interface{} {
	return &event.StepQueue{}
}

// * This is the handler that will actually execute the primitive
func (h StepQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.StepQueue)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.StepQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.StepQueue")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	if cmd.StepRetry != nil {
		ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(cmd.Event.ExecutionID, cmd.PipelineExecutionID)
		if err != nil {
			slog.Error("pipeline_plan: Error loading pipeline execution", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
			return nil
		}

		pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

		evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
		if err != nil {
			slog.Error("Error building eval context", "error", err)
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
		}

		stepDefn := pipelineDefn.GetStep(cmd.StepName)

		retryConfig, diags := stepDefn.GetRetryConfig(evalContext, false)
		if len(diags) > 0 {
			slog.Error("Error getting retry config", "diags", diags)
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, error_helpers.HclDiagsToError(stepDefn.GetName(), diags))))
		}

		if retryConfig != nil {

			// The first retry is the second attempt. StepRetry.Count is not an index, but it's a count of how many times we have retried
			// this step.
			//
			// So ... to calculate the backoff, we need to add 1 to the count because the 1st retry is the 2nd count.
			duration := retryConfig.CalculateBackoff(cmd.StepRetry.Count + 1)

			slog.Info("Delaying step start for", "duration", duration, "stepName", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID)
			start := time.Now().UTC()
			time.Sleep(duration)
			finish := time.Now().UTC()

			slog.Info("Delaying step start complete", "duration", duration, "stepName", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID, "start", start, "finish", finish)

			e, err := event.NewStepQueued(event.ForStepQueue(cmd))
			if err != nil {
				slog.Error("Error creating step queued event", "error", err)
				err = h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
				if err != nil {
					slog.Error("Error publishing pipeline failed event", "error", err)
				}
			}
			err = h.EventBus.Publish(ctx, e)
			if err != nil {
				slog.Error("Error publishing step queued event", "error", err)
				err = h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
				if err != nil {
					slog.Error("Error publishing pipeline failed event", "error", err)
				}
			}

			return nil
		}
	}

	e, err := event.NewStepQueued(event.ForStepQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
