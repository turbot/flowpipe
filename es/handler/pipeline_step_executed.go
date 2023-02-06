package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStepExecuted EventHandler

func (h PipelineStepExecuted) HandlerName() string {
	return "handler.pipeline_step_executed"
}

func (PipelineStepExecuted) NewEvent() interface{} {
	return &event.PipelineStepExecuted{}
}

func (h PipelineStepExecuted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStepExecuted)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	cmd := event.PipelinePlan{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now().UTC(),
		StackID:   e.StackID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
