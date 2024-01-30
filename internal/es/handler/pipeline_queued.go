package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return execution.PipelineQueuedEvent.HandlerName()
}

func (PipelineQueued) NewEvent() interface{} {
	return &event.PipelineQueued{}
}

// Path from here:
// * PipelineQueued -> PipelineLoad command -> PipelineLoaded event handler
func (h PipelineQueued) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.PipelineQueued)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineQueued")
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
		slog.Error("pipeline_queued: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
			return nil
		}

	}

	cmd, err := event.NewPipelineLoad(event.ForPipelineQueued(evt))
	if err != nil {
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
			return nil
		}
	}
	// Make sure we release the planner mutex here, otherwise we'll create a deadlock
	// when the step start command hanlder tries to acquire the mutex to "finish" the step
	plannerMutex.Unlock()
	plannerMutex = nil

	go func() {
		err := execution.GetPipelineSemaphore(pipelineDefn)
		if err != nil {
			err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(evt, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
				return
			}
		}

		plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
		plannerMutex.Lock()

		defer func() {
			if plannerMutex != nil {
				plannerMutex.Unlock()
			}
		}()

		err = h.CommandBus.Send(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
			return
		}

	}()
	return nil
}
