package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
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
	e := ei.(*event.PipelineCanceled)
	fplog.Logger(ctx).Info("pipeline_canceled", "error", e)
	return nil
}
