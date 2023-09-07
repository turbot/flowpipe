package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type PipelineStepStarted EventHandler

func (h PipelineStepStarted) HandlerName() string {
	return "handler.pipeline_step_started"
}

func (PipelineStepStarted) NewEvent() interface{} {
	return &event.PipelineStepStarted{}
}

// This handler only handle with a single event type: pipeline step started (if we want to start a new child pipeline)
func (h PipelineStepStarted) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelineStepStarted)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineStepStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineStepStarted")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}

	stepDefn, err := ex.StepDefinition(e.PipelineExecutionID, e.StepExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}

	switch stepDefn.GetType() {
	case schema.BlockTypePipelineStepPipeline:
		cmd, err := event.NewPipelineQueue(event.ForPipelineStepStartedToPipelineQueue(e))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, &cmd)
	default:
		err := perr.BadRequestWithMessage("step type cannot be started: " + stepDefn.GetType())
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}
}
