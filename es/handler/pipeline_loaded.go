package handler

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/es/event"
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

	fmt.Println()
	fmt.Println("I am handling the Pipeline Loaded event")
	fmt.Println()
	cmd, err := event.NewPipelineStart(event.ForPipelineLoaded(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineLoadedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
