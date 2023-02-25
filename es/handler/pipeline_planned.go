package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	// PRE: The planner has told us what to run next, our job is to schedule it

	s, err := state.NewState(ctx, e.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := command.PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	for _, stepIndex := range e.NextStepIndexes {
		cmd := event.PipelineStepExecute{
			RunID:     e.RunID,
			SpanID:    e.SpanID,
			CreatedAt: time.Now().UTC(),
			StackID:   e.StackID + "." + xid.New().String(),
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
		complete := true
		for _, stepStatus := range s.PipelineStepStatus {
			if stepStatus != "completed" {
				complete = false
				break
			}
		}
		if complete {
			cmd := event.PipelineFinish{
				RunID:     e.RunID,
				SpanID:    e.SpanID,
				CreatedAt: time.Now().UTC(),
				StackID:   e.StackID,
			}
			return h.CommandBus.Send(ctx, &cmd)
		}
	}

	return nil
}
