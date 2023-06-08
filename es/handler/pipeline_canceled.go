package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineCanceled EventHandler

func (h PipelineCanceled) HandlerName() string {
	return "handler.pipeline_canceled"
}

func (PipelineCanceled) NewEvent() interface{} {
	return &event.PipelineCanceled{}
}

func (h PipelineCanceled) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineCanceled)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineCanceled", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineCanceled")
	}

	logger.Info("[4] pipeline_canceled event handler", "event", e)
	return nil
}
