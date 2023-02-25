package command

import (
	"context"
	"fmt"
	"time"

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

	s, err := state.NewState(ctx, cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	e := event.PipelinePlanned{
		RunID:           cmd.RunID,
		SpanID:          cmd.SpanID,
		CreatedAt:       time.Now().UTC(),
		NextStepIndexes: []int{},
	}

	// Plan steps for execution, but only if their dependencies have been met.
	for i, step := range defn.Steps {
		// If the step is already planned, running, etc, then skip it.
		if s.PipelineStepStatus[i] != "" {
			continue
		}
		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range step.DependsOn {
			if s.PipelineStepStatus[dep] != "completed" {
				dependendenciesMet = false
				break
			}
		}
		if !dependendenciesMet {
			continue
		}
		// Plan to run the step.
		e.NextStepIndexes = append(e.NextStepIndexes, i)
	}

	return h.EventBus.Publish(ctx, &e)
}
