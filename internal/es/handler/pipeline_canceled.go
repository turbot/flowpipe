package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
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

	err = ex.SaveToFile()
	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	event.ReleaseEventLogMutex(evt.Event.ExecutionID)
	execution.CompletePipelineExecutionStepSemaphore(evt.PipelineExecutionID)
	err = execution.ReleasePipelineSemaphore(pipelineDefn)
	if err != nil {
		slog.Error("pipeline_finished: Error releasing pipeline semaphore", "error", err)
		return nil
	}

	return nil
}
