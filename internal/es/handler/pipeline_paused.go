package handler

import (
	"context"

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
	event.ReleaseEventLogMutex(evt.Event.ExecutionID)

	return nil
}
