package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStarted EventHandler

func (h PipelineStarted) HandlerName() string {
	return "handler.pipeline_started"
}

func (PipelineStarted) NewEvent() interface{} {
	return &event.PipelineStarted{}
}

func (h PipelineStarted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStarted)

	cmd := event.PipelinePlan{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, &cmd)
}
