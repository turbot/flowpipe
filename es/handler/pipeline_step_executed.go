package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStepExecuted EventHandler

func (h PipelineStepExecuted) HandlerName() string {
	return "handler.executed"
}

func (PipelineStepExecuted) NewEvent() interface{} {
	return &event.PipelineStepExecuted{}
}

func (h PipelineStepExecuted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStepExecuted)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	//Not sure what this was doing, but it created infinite loops
	cmd := event.PipelinePlan{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
		StackID:   e.StackID,
	}

	return h.CommandBus.Send(ctx, &cmd)
}

/*

	s, err := state.NewState(e.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// TODO - pipeline name needs to be read from the state
	defn, err := command.PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	nextStepIndex := s.Stack[e.StackID].StepIndex + 1

	var nextStackID string

	lastPartIndex := strings.LastIndex(e.StackID, ".")
	if lastPartIndex == -1 {
		nextStackID = e.StackID + "." + xid.New().String()
	} else {
		nextStackID = e.StackID[:strings.LastIndex(e.StackID, ".")+1] + xid.New().String()
	}

	// Run the next step
	cmd := event.PipelineStepExecute{
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

*/
