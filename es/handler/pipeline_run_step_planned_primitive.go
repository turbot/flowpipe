package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepPrimitivePlanned EventHandler

func (h PipelineRunStepPrimitivePlanned) HandlerName() string {
	return "handler.pipeline_run_step_planned_primitive"
}

func (PipelineRunStepPrimitivePlanned) NewEvent() interface{} {
	return &event.PipelineRunStepPrimitivePlanned{}
}

func (h PipelineRunStepPrimitivePlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepPrimitivePlanned)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	// We have another step to run
	cmd := &command.PipelineRunStepPrimitiveExecute{
		RunID:     e.RunID,
		Primitive: e.Primitive,
		Input:     e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
