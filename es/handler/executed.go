package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"

	"github.com/turbot/steampipe-pipelines/es/state"
)

type Executed EventHandler

func (h Executed) HandlerName() string {
	return "handler.executed"
}

func (Executed) NewEvent() interface{} {
	return &event.Executed{}
}

func (h Executed) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Executed)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	/*

			//Not sure what this was doing, but it created infinite loops
			cmd := command.PipelinePlan{
				SpanID:   e.SpanID,
				StackID: e.StackID,
			}

			return h.CommandBus.Send(ctx, &cmd)
		}

	*/

	s, err := state.NewState(e.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// TODO - pipeline name needs to be read from the state
	//defn, err := PipelineDefinition(s.PipelineName)
	defn, err := command.PipelineDefinition("my_pipeline_0")
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	nextStepIndex := s.Stack[e.StackID].StepIndex + 1

	if nextStepIndex >= len(defn.Steps) {
		return nil
		/*
			// Nothing to do!
			cmd := &command.Stop{
				SpanID: e.SpanID,
			}
			return h.CommandBus.Send(ctx, cmd)
		*/
	}

	var nextStackID string

	lastPartIndex := strings.LastIndex(e.StackID, ".")
	if lastPartIndex == -1 {
		nextStackID = e.StackID + "." + xid.New().String()
	} else {
		nextStackID = e.StackID[:strings.LastIndex(e.StackID, ".")+1] + xid.New().String()
	}

	// Run the next step
	cmd := event.Execute{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
		StackID:   nextStackID,
		//PipelineName: s.PipelineName,
		StepIndex: nextStepIndex,
		Input:     defn.Steps[nextStepIndex].Input,
	}

	return h.CommandBus.Send(ctx, &cmd)
}
