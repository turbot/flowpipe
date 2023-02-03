package handler

import (
	"context"
	"fmt"
	"time"

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

	cmd := event.PipelinePlan{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
		StackID:   e.StackID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
