package handler

import (
	"context"

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

	if len(e.Event.StackIDs) > 1 {
		// This is a child pipeline, so trigger the planner for the parent
		// pipeline.
		cmd := event.PipelinePlan{
			Event: event.NewParentEvent(event.NewParentEvent(e.Event)),
		}
		/*
			cmd := event.PipelineStepFinish{
				Event:     event.NewParentEvent(e.Event),
				StepIndex: cmd.StepIndex,
			}
		*/
		return h.CommandBus.Send(ctx, &cmd)
	}

	return nil
}
