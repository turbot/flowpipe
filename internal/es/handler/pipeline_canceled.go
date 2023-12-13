package handler

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
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

	execution.ServerOutput(fmt.Sprintf("[%s] Pipeline cancelled", e.Event.ExecutionID))

	event.ReleaseEventLogMutex(e.Event.ExecutionID)

	return nil
}
