package handler

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueued EventHandler

func (h StepQueued) HandlerName() string {
	return execution.StepQueuedEvent.HandlerName()
}

func (StepQueued) NewEvent() interface{} {
	return &event.StepQueued{}
}

func (h StepQueued) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.StepQueued)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepQueued")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	_, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("step_queued: Error loading pipeline execution", "error", err)
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepQueuedToPipelineFail(evt, err)))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return nil
	}

	stepDefn := pipelineDefn.GetStep(evt.StepName)

	// Step has been queued (but not yet started), so here we just need to start the step
	// the code should be the same as the pipeline_planned event handler
	evt.StepType = stepDefn.GetType()
	cmd, err := event.NewStepStartFromStepQueued(evt)

	if err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepQueuedToPipelineFail(evt, err)))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return nil
	}

	// Make sure we release the planner mutex here, otherwise we'll create a deadlock
	// when the step start command hanlder tries to acquire the mutex to "finish" the step
	plannerMutex.Unlock()
	plannerMutex = nil
	execution.GetStepTypeSemaphore(evt.StepType)

	plannerMutex = event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()

	if err := h.CommandBus.Send(ctx, cmd); err != nil {
		err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForStepQueuedToPipelineFail(evt, err)))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return nil
	}

	return nil
}
