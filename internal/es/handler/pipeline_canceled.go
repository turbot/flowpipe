package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/pipe-fittings/perr"
	"log/slog"
)

type PipelineCanceled EventHandler

func (h PipelineCanceled) HandlerName() string {
	return execution.PipelineCanceledEvent.HandlerName()
}

func (PipelineCanceled) NewEvent() interface{} {
	return &event.PipelineCanceled{}
}

func (h PipelineCanceled) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineCanceled)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineCanceled", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineCanceled")
	}

	eventStoreFilePath := filepaths.EventStoreFilePath(e.Event.ExecutionID)
	err := sanitize.Instance.SanitizeFile(eventStoreFilePath)
	if err != nil {
		slog.Error("Failed to sanitize file", "eventStoreFilePath", eventStoreFilePath)
	}

	event.ReleaseEventLogMutex(e.Event.ExecutionID)

	return nil
}
