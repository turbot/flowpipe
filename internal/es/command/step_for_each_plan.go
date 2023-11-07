package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type StepForEachPlanHandler CommandHandler

var stepForEachPlan = event.StepForEachPlan{}

func (h StepForEachPlanHandler) HandlerName() string {
	return stepForEachPlan.HandlerName()
}

func (h StepForEachPlanHandler) NewCommand() interface{} {
	return &event.StepForEachPlan{}
}

func (h StepForEachPlanHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	_, ok := c.(*event.StepForEachPlan)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.StepForEachPlan", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.StepForEachPlan")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[e.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pe.IsCanceled() || pe.IsPaused() || pe.IsFinishing() || pe.IsFinished() {
		return nil
	}

	return nil
}
