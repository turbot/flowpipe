package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
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
	cmd, err := event.NewPipelinePlan(event.ForPipelineStarted(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStartedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
