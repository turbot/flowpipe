package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepPrimitivePlanned EventHandler

func (h PipelineRunStepPrimitivePlanned) HandlerName() string {
	return "handler.pipeline_run_step_primitive_planned"
}

func (PipelineRunStepPrimitivePlanned) NewEvent() interface{} {
	return &event.PipelineRunStepPrimitivePlanned{}
}

func (h PipelineRunStepPrimitivePlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepPrimitivePlanned)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	// We have another step to run
	cmd := &event.PipelineRunStepPrimitiveExecute{
		SpanID:    e.SpanID,
		Primitive: e.Primitive,
		Input:     e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
