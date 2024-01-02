package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type StepFinished EventHandler

func (h StepFinished) HandlerName() string {
	return execution.StepFinishedEvent.HandlerName()
}

func (StepFinished) NewEvent() interface{} {
	return &event.StepFinished{}
}

// This is the generic step finish event handler that is fired by the step_start command
//
// Do not confuse this with pipeline_step_finish **command** which is raised when a child pipeline has finished
func (h StepFinished) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.StepFinished)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepFinished")
	}

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("step_finished: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	pex := ex.PipelineExecutions[evt.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	stepExecution := pex.StepExecutions[evt.StepExecutionID]
	stepName := stepExecution.Name

	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
	}

	stepDefn := pipelineDefn.GetStep(stepName)
	if stepDefn == nil {
		slog.Error("step not found", "step_name", stepName)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, perr.BadRequestWithMessage("step not found"))))
	}

	// Check if we are in a retry block
	if evt.StepRetry != nil && !evt.StepRetry.RetryCompleted {
		if output.IsServerMode {
			step := types.NewServerOutputStepExecution(
				types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
				evt.Event.ExecutionID,
				pipelineDefn.PipelineName,
				stepName,
				stepDefn.GetType(),
				"retrying",
			)
			output.RenderServerOutput(ctx, step)
		}
		cmd := event.NewStepQueueFromPipelineStepFinishedForRetry(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	} else if evt.StepRetry != nil && evt.StepRetry.RetryCompleted {
		// this means we have an error BUT the retry has been exhausted, run the planner
		if output.IsServerMode {
			step := types.NewServerOutputStepExecution(
				types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
				evt.Event.ExecutionID,
				pipelineDefn.PipelineName,
				stepName,
				stepDefn.GetType(),
				"failed",
			)
			step.Output = evt.StepOutput
			if !helpers.IsNil(evt.Output) && len(evt.Output.Errors) > 0 {
				step.Errors = evt.Output.Errors
			}
			output.RenderServerOutput(ctx, step)
		}
		cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(evt))
		if err != nil {
			slog.Error("error creating pipeline_plan command", "error", err)
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// First thing first .. before we run the planner (either pipeline plan or step for each plan),
	// check if we are in a loop. If we are in a loop start the next loop
	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	if loopBlock != nil && evt.StepLoop != nil && !evt.StepLoop.LoopCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForLoop(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	// If the step is a for each step, run the for each planner, not the pipeline planner
	if !helpers.IsNil(stepDefn.GetForEach()) {
		cmd := event.NewStepForEachPlanFromPipelineStepFinished(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	if output.IsServerMode {
		step := types.NewServerOutputStepExecution(
			types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
			evt.Event.ExecutionID,
			pipelineDefn.PipelineName,
			stepName,
			stepDefn.GetType(),
			"finished",
		)
		step.Output = evt.StepOutput
		if !helpers.IsNil(evt.Output) && len(evt.Output.Errors) > 0 {
			step.Errors = evt.Output.Errors
			step.Status = "failed"
		}
		output.RenderServerOutput(ctx, step)
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(evt))
	if err != nil {
		slog.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
	}

	return h.CommandBus.Send(ctx, cmd)
}
