package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStepExecuted EventHandler

func (h PipelineStepExecuted) HandlerName() string {
	return "handler.executed"
}

func (PipelineStepExecuted) NewEvent() interface{} {
	return &event.PipelineStepExecuted{}
}

func (h PipelineStepExecuted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStepExecuted)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	//Not sure what this was doing, but it created infinite loops
	cmd := event.PipelinePlan{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now().UTC(),
		StackID:   e.StackID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
