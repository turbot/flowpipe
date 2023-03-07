package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineLoaded EventHandler

func (h PipelineLoaded) HandlerName() string {
	return "handler.pipeline_loaded"
}

func (PipelineLoaded) NewEvent() interface{} {
	return &event.PipelineLoaded{}
}

func (h PipelineLoaded) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineLoaded)

	cmd := &event.PipelineStart{
		Event:        event.NewFlowEvent(e.Event),
		PipelineName: e.Pipeline.Name,
		//Input:        e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
