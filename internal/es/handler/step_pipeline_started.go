package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type StepPipelineStarted EventHandler

func (h StepPipelineStarted) HandlerName() string {
	return execution.StepPipelineStartedEvent.HandlerName()
}

func (StepPipelineStarted) NewEvent() interface{} {
	return &event.StepPipelineStarted{}
}

// *
// * This handler only handle with a single event type: pipeline step started (if we want to start a new child pipeline)
// *
func (h StepPipelineStarted) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.StepPipelineStarted)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.StepPipelineStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepPipelineStarted")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(e, err)))
	}

	stepDefn, err := ex.StepDefinition(e.PipelineExecutionID, e.StepExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(e, err)))
	}

	switch stepDefn.GetType() {
	case schema.BlockTypePipelineStepPipeline:
		cmd, err := event.NewPipelineQueue(event.ForPipelineStepStartedToPipelineQueue(e))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, cmd)
	default:
		err := perr.BadRequestWithMessage("step type cannot be started: " + stepDefn.GetType())
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(e, err)))
	}
}
