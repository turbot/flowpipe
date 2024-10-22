package handler

import (
	"context"
	"log/slog"
	"slices"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineCanceled EventHandler

func (h PipelineCanceled) HandlerName() string {
	return execution.PipelineCanceledEvent.HandlerName()
}

func (PipelineCanceled) NewEvent() interface{} {
	return &event.PipelineCanceled{}
}

func (h PipelineCanceled) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.PipelineCanceled)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineCanceled", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineCanceled")
	}

	err := store.UpdatePipelineState(evt.Event.ExecutionID, "cancelled")
	if err != nil {
		slog.Error("pipeline_cancelled: Error updating pipeline state", "error", err)
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	_, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_cancelled: Error loading pipeline execution", "error", err)
		return err
	}

	ex, err := execution.GetExecution(evt.Event.ExecutionID)
	if err != nil {
		slog.Error("pipeline_finished: Error loading pipeline execution", "error", err)
		return err
	}

	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	pipelineCompletionHandler(evt.Event.ExecutionID, evt.PipelineExecutionID, pipelineDefn, ex.PipelineExecutions[evt.PipelineExecutionID].StepExecutions)

	// raise execution plan command if this pipeline is in the root pipeline list
	if slices.Contains(ex.RootPipelines, evt.PipelineExecutionID) {
		cmd := event.ExecutionPlanFromPipelineCancelled(evt)
		err = h.CommandBus.Send(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
	}
	return nil
}
