package handler

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
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
	evt, ok := ei.(*event.StepPipelineStarted)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepPipelineStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepPipelineStarted")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, err := execution.GetExecution(evt.Event.ExecutionID)
	if err != nil {
		slog.Error("step_pipeline_started: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	stepDefn, err := ex.StepDefinition(evt.PipelineExecutionID, evt.StepExecutionID)
	if err != nil {
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
	}

	switch stepDefn.GetType() {
	case schema.BlockTypePipelineStepPipeline:
		cmd, err := event.NewPipelineQueue(event.ForPipelineStepStartedToPipelineQueue(evt))
		if err != nil {
			err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(evt, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
		}

		return h.CommandBus.Send(ctx, cmd)
	default:

		err := perr.BadRequestWithMessage("step type cannot be started: " + stepDefn.GetType())
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepPipelineStartedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
	}

	return nil
}
