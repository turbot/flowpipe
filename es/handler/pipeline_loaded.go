package handler

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	cmd := &event.PipelineStart{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now().UTC(),
		//StackID:      e.StackID,
		PipelineName: e.PipelineName,
		//Input:        e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
