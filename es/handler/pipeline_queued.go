package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return "handler.pipeline_queued"
}

func (PipelineQueued) NewEvent() interface{} {
	return &event.PipelineQueued{}
}

func (h PipelineQueued) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineQueued)

	cmd := &event.PipelineLoad{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, cmd)
}
