package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelinePaused EventHandler

func (h PipelinePaused) HandlerName() string {
	return "handler.pipeline_paused"
}

func (PipelinePaused) NewEvent() interface{} {
	return &event.PipelinePaused{}
}

func (h PipelinePaused) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelinePaused)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePaused", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelinePaused")
	}

	logger.Info("[8] pipeline_paused event handler", "eventExecutionID", e.Event.ExecutionID)
	return nil
}
