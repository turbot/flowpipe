package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineFailed EventHandler

func (h PipelineFailed) HandlerName() string {
	return "handler.pipeline_failed"
}

func (PipelineFailed) NewEvent() interface{} {
	return &event.PipelineFailed{}
}

func (h PipelineFailed) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineFailed)
	fplog.Logger(ctx).Error("pipeline_failed (4)", "error", e)
	return nil
}
