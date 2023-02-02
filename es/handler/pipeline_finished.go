package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return "handler.pipeline_finished"
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineFinished)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	// TODO - this should pop off the stack, not just straight to the top
	cmd := command.Finish{
		RunID: e.RunID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
