package handler

import (
	"context"

	"github.com/pkg/errors"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineStepStarted EventHandler

func (h PipelineStepStarted) HandlerName() string {
	return "handler.pipeline_step_started"
}

func (PipelineStepStarted) NewEvent() interface{} {
	return &event.PipelineStepStarted{}
}

func (h PipelineStepStarted) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelineStepStarted)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineStepStarted", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineStepStarted")
	}

	logger.Info("[13] pipeline_step_started event handler", "executionID", e.Event.ExecutionID)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}

	stepDefn, err := ex.StepDefinition(e.StepExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}

	switch stepDefn.Type {
	case "pipeline":
		cmd, err := event.NewPipelineQueue(event.ForPipelineStepStartedToPipelineQueue(e))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, &cmd)
	case "sleep":
		// TODO - implement
	default:
		err := errors.Errorf("step type cannot be started: %s", stepDefn.Type)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepStartedToPipelineFail(e, err)))
	}

	return nil
}
