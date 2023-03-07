package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type PipelineStepStarted EventHandler

func (h PipelineStepStarted) HandlerName() string {
	return "handler.pipeline_step_started"
}

func (PipelineStepStarted) NewEvent() interface{} {
	return &event.PipelineStepStarted{}
}

func (h PipelineStepStarted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStepStarted)

	s, err := state.NewState(ctx, e.Event)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := command.PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	step := defn.Steps[e.StepIndex]

	if step.Type == "pipeline" {
		cmd := event.PipelineQueue{
			Event: event.NewChildEvent(e.Event),
			Name:  step.Input["name"].(string),
		}
		return h.CommandBus.Send(ctx, &cmd)
	}

	return nil
}
