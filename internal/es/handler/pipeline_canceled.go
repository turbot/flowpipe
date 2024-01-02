package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"

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

	ex, err := execution.GetExecution(evt.Event.ExecutionID)
	if err != nil {
		slog.Error("pipeline_finished: Error loading pipeline execution", "error", err)
		return err
	}

	ex.Lock.Lock()
	defer ex.Lock.Unlock()

	err = ex.SaveToFile()
	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
			evt.Event.ExecutionID, "", "cancelled")
		output.RenderServerOutput(ctx, p)
	}

	event.ReleaseEventLogMutex(evt.Event.ExecutionID)

	return nil
}
