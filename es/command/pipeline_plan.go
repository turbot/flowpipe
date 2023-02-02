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

type PipelinePlan struct {
	RunID   string `json:"run_id"`
	StackID string `json:"stack_id"`
}

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return "command.pipeline_plan"
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &PipelinePlan{}
}

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*PipelinePlan)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	s, err := state.NewState(cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// Load the pipeline definition
	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	nextStepIndex := s.Stack[cmd.StackID].StepIndex + 1

	if nextStepIndex >= len(defn.Steps) {
		// Nothing to do!
		e := event.PipelineFinished{
			RunID:   cmd.RunID,
			StackID: cmd.StackID,
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

	e := event.PipelinePlanned{
		RunID:     cmd.RunID,
		StackID:   nextStackID,
		Timestamp: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
