package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PipelineStepFinished EventHandler

func (h PipelineStepFinished) HandlerName() string {
	return "handler.pipeline_step_finished"
}

func (PipelineStepFinished) NewEvent() interface{} {
	return &event.PipelineStepFinished{}
}

func (h PipelineStepFinished) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineStepFinished)
	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.PipelineStepFinished", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineStepFinished")
	}

	logger := fplog.Logger(ctx)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		logger.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[e.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pe.IsCanceled() || pe.IsPaused() || pe.IsFinishing() || pe.IsFinished() {
		return nil
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(e))
	if err != nil {
		logger.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(e, err)))
	}

	return h.CommandBus.Send(ctx, cmd)
}
