package handler

import (
	"context"

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

	//e := ei.(*event.Started)

	// Note: The mod is now listening for trigger events. It is stopped by a
	// Ctrl-C handler hooked to the Stop command.

	return nil
}

/*

func (h Started) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Started)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	s, err := state.NewState(e.SpanID)
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
			SpanID: e.SpanID,
		}
		return h.CommandBus.Send(ctx, &cmd)
	}

	// Run the first step
	cmd := event.Execute{
		SpanID:        e.SpanID,
		StackID:      e.StackID + "." + xid.New().String(),
		PipelineName: s.PipelineName,
		StepIndex:    0,
		Input:        defn.Steps[0].Input,
	}

	return h.CommandBus.Send(ctx, &cmd)
}

*/
