package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineCanceled EventHandler

var pipelineCanceled = event.PipelineCanceled{}

func (h PipelineCanceled) HandlerName() string {
	return pipelineCanceled.HandlerName()
}

func (PipelineCanceled) NewEvent() interface{} {
	return &event.PipelineCanceled{}
}

func (h PipelineCanceled) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineCanceled)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineCanceled", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineCanceled")
	}

	logger.Info("[4] pipeline_canceled event handler", "event", e)
	return nil
}
