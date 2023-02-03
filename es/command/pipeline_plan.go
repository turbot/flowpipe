package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return "command.pipeline_plan"
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &event.PipelinePlan{}
}

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelinePlan)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	s, err := state.NewState(cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// Load the pipeline definition
	// TODO - definition be based off the load phase
	// TODO - pipeline name needs to be read from the state
	//defn, err := PipelineDefinition(s.PipelineName)
	defn, err := PipelineDefinition("my_pipeline_0")
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	nextStepIndex := s.Stack[cmd.StackID].StepIndex + 1

	if nextStepIndex >= len(defn.Steps) {
		// Nothing to do!
		e := event.PipelineFinished{
			RunID:     cmd.RunID,
			SpanID:    cmd.SpanID,
			CreatedAt: time.Now(),
			StackID:   cmd.StackID,
		}
		return h.EventBus.Publish(ctx, &e)
	}

	var nextStackID string

	lastPartIndex := strings.LastIndex(cmd.StackID, ".")
	if lastPartIndex == -1 {
		nextStackID = cmd.StackID + "." + xid.New().String()
	} else {
		nextStackID = cmd.StackID[:strings.LastIndex(cmd.StackID, ".")+1] + xid.New().String()
	}

	// Send a planned event with the information about which step to run next.
	e := event.PipelinePlanned{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		StackID:   nextStackID,
	}

	return h.EventBus.Publish(ctx, &e)
}
