package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
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
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.StepFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepFinished")
	}

	logger := fplog.Logger(ctx)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		logger.Error("error creating pipeline_plan command", "error", err)
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
		logger.Error("step not found", "step_name", stepName)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, perr.BadRequestWithMessage("step not found"))))
	}

	if e.Output != nil && (len(e.Output.Errors) > 0 || e.Output.Status == "failed") {
		h.handleError(ctx, e, stepDefn)
		return nil
	}

	// First thing first .. before we run the planner (either pipeline plan or step for each plan),
	// check if we are in a loop. If we are in a loop start the next loop
	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	if loopBlock != nil && e.StepLoop != nil && !e.StepLoop.LoopCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForLoop(e, stepName)
		if err != nil {
			err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
			if err != nil {
				logger.Error("Error publishing event", "error", err)
			}
			return nil
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	if !helpers.IsNil(stepDefn.GetForEach()) {
		cmd := event.NewStepForEachPlanFromPipelineStepFinished(e, stepName)

		return h.CommandBus.Send(ctx, cmd)
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
	if err != nil {
		logger.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	return h.CommandBus.Send(ctx, cmd)
}

// errors are handled in the following order:
//
// throw, in the order that they appear
// retry
// error
func (h StepFinished) handleError(ctx context.Context, e *event.StepFinished, stepDefn modconfig.PipelineStep) {
	// we have error, check the if there's a retry block

}
