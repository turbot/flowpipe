package handler

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	cmd := &event.PipelineLoad{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now().UTC(),
	}

	return h.CommandBus.Send(ctx, cmd)
}
