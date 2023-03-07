package command

import (
	"context"
	"fmt"

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

	s, err := state.NewState(ctx, cmd.Event)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	fmt.Println(cmd.Event.StackIDs)
	fmt.Println(s)

	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	e := event.PipelinePlanned{
		Event:           event.NewFlowEvent(cmd.Event),
		NextStepIndexes: []int{},
	}

	//lastStackID := cmd.Event.LastStackID()
	lastStackID := cmd.Event.StackIDs[len(cmd.Event.StackIDs)-1]
	stack := s.Stacks[lastStackID]
	//s.Stacks[lastStackID].StepStatus[et.StepIndex] = "started"

	// Plan steps for execution, but only if their dependencies have been met.
	for i, step := range defn.Steps {
		// If the step is already planned, running, etc, then skip it.
		if stack.StepStatus[i] != "" {
			//if s.PipelineStepStatus[i] != "" {
			continue
		}
		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range step.DependsOn {
			if stack.StepStatus[dep] != "finished" {
				//if s.PipelineStepStatus[dep] != "finished" {
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
