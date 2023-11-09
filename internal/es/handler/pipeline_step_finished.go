package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type PipelineStepFinished EventHandler

var pipelineStepFinished = event.PipelineStepFinished{}

func (h PipelineStepFinished) HandlerName() string {
	return pipelineStepFinished.HandlerName()
}

func (PipelineStepFinished) NewEvent() interface{} {
	return &event.PipelineStepFinished{}
}

func (h PipelineStepFinished) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineStepFinished)
	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.PipelineStepFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineStepFinished")
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

	// First thing first .. before we run the planner (either pipeline plan or step for each plan),
	// check if we are in a loop. If we are in a loop start the next loop
	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	if loopBlock != nil && e.StepLoop != nil && !e.StepLoop.LoopCompleted {
		
		cmd := event.NewPipelineStepQueueFromPipelineStepFinishedForLoop(e, stepName)
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
