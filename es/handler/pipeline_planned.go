package handler

import (
	"context"
	"fmt"

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

	s, err := state.NewState(e.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// Load the pipeline definition
	defn, err := command.PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	if len(defn.Steps) <= 0 {
		// Nothing to do!
		// TODO - should be PipelineFinish
		cmd := event.PipelineFinish{
			RunID:   e.RunID,
			StackID: e.StackID,
		}
		return h.CommandBus.Send(ctx, &cmd)
	}

	// Run the first step
	cmd := event.Execute{
		RunID:        e.RunID,
		StackID:      e.StackID + "." + xid.New().String(),
		PipelineName: s.PipelineName,
		StepIndex:    0,
		Input:        defn.Steps[0].Input,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
