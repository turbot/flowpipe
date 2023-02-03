package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Started EventHandler

func (h Started) HandlerName() string {
	return "handler.started"
}

func (Started) NewEvent() interface{} {
	return &event.Started{}
}

func (h Started) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Started)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	// TODO - Now we are ready to receive events, we should turn on the
	// handlers.

	/*

		TESTING - just run a pipeline

		cmd := &event.PipelineStart{
			RunID:        e.RunID,
			StackID:      e.StackID,
			PipelineName: e.PipelineName,
			Input:        e.Input,
		}

		return h.CommandBus.Send(ctx, cmd)

	*/

	return nil

}

/*

func (h Started) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Started)

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
		cmd := command.Finish{
			RunID: e.RunID,
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

*/
