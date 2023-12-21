package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
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
	e, ok := ei.(*event.StepFinished)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepFinished")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		slog.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	// Convenience
	pex := ex.PipelineExecutions[e.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	stepExecution := pex.StepExecutions[e.StepExecutionID]
	stepName := stepExecution.Name

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	stepDefn := pipelineDefn.GetStep(stepName)
	if stepDefn == nil {
		slog.Error("step not found", "step_name", stepName)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, perr.BadRequestWithMessage("step not found"))))
	}

	// Check if we are in a retry block
	if e.StepRetry != nil && !e.StepRetry.RetryCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForRetry(e, stepName)
		return h.CommandBus.Send(ctx, cmd)
	} else if e.StepRetry != nil && e.StepRetry.RetryCompleted {
		// this means we have an error BUT the retry has been exhausted, run the planner
		cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
		if err != nil {
			slog.Error("error creating pipeline_plan command", "error", err)
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// First thing first .. before we run the planner (either pipeline plan or step for each plan),
	// check if we are in a loop. If we are in a loop start the next loop
	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	if loopBlock != nil && e.StepLoop != nil && !e.StepLoop.LoopCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForLoop(e, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	// If the step is a for each step, run the for each planner, not the pipeline planner
	if !helpers.IsNil(stepDefn.GetForEach()) {
		cmd := event.NewStepForEachPlanFromPipelineStepFinished(e, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	// execution.ServerOutput(fmt.Sprintf("[%s] Step %s finished", e.Event.ExecutionID, stepDefn.GetName()))

	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
	if err != nil {
		slog.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	return h.CommandBus.Send(ctx, cmd)
}
