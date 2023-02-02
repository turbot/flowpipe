package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	cmd := command.PipelinePlan{
		RunID:   e.RunID,
		StackID: e.StackID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
