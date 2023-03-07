package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type PipelinePlanned EventHandler

func (h PipelinePlanned) HandlerName() string {
	return "handler.pipeline_planned"
}

func (PipelinePlanned) NewEvent() interface{} {
	return &event.PipelinePlanned{}
}

func (h PipelinePlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelinePlanned)

	// PRE: The planner has told us what to run next, our job is to schedule it

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

	// Start execution of any next steps from the plan.
	for _, stepIndex := range e.NextStepIndexes {
		cmd := event.PipelineStepStart{
			Event: event.NewChildEvent(e.Event),
			//Event:     event.NewFlowEvent(e.Event),
			StepIndex: stepIndex,
			Input:     defn.Steps[stepIndex].Input,
		}
		if err := h.CommandBus.Send(ctx, &cmd); err != nil {
			return err
		}
	}

	// If there are no more steps, and all running steps are complete, then the
	// pipeline is complete.
	if len(e.NextStepIndexes) == 0 {

		lastStackID := e.Event.StackIDs[len(e.Event.StackIDs)-1]
		//lastStackID := e.Event.LastStackID()
		stack := s.Stacks[lastStackID]

		complete := true
		for stepID := range defn.Steps {
			if stack.StepStatus[stepID] != "finished" {
				//if s.PipelineStepStatus[stepID] != "finished" {
				complete = false
				break
			}
		}
		if complete {
			cmd := event.PipelineFinish{
				Event: event.NewFlowEvent(e.Event),
			}
			return h.CommandBus.Send(ctx, &cmd)
		}
	}

	return nil
}
