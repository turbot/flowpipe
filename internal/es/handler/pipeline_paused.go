package handler

import (
	"context"
	"slices"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelinePaused EventHandler

func (h PipelinePaused) HandlerName() string {
	return execution.PipelinePausedEvent.HandlerName()
}

func (PipelinePaused) NewEvent() interface{} {
	return &event.PipelinePaused{}
}

func (h PipelinePaused) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.PipelinePaused)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelinePaused", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePaused")
	}

	slog.Info("PipelinePaused event received", "execution_id", evt.Event.ExecutionID, "pipeline_execution_id", evt.PipelineExecutionID)

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, _, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_finished: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.PipelineFailFromPipelinePaused(evt, err))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	// raise execution plan command if this pipeline is in the root pipeline list
	if slices.Contains(ex.RootPipelines, evt.PipelineExecutionID) {
		cmd := event.ExecutionPlanFromPipelinePaused(evt)
		err = h.CommandBus.Send(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
	}

	return nil
}
