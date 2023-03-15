package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
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

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePlanToPipelineFailed(cmd, err)))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePlanToPipelineFailed(cmd, err)))
	}

	e, err := event.NewPipelinePlanned(event.ForPipelinePlan(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePlanToPipelineFailed(cmd, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	// Each defined step in the pipeline can be in a few states:
	// - dependencies not met
	// - queued
	// - started
	// - finished
	// - failed
	//
	// Notably each step may also have multiple executions (e.g. in a for
	// loop). So, we need to track the overall status of the step separately
	// from the status of each execution.
	for _, step := range defn.Steps {

		if pe.IsStepInitialized(step.Name) {
			continue
		}

		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range step.DependsOn {
			// Cannot depend on yourself
			if step.Name == dep {
				// TODO - issue a warning?
				continue
			}
			// Ignore invalid dependencies
			if _, ok := defn.Steps[dep]; !ok {
				// TODO - issue a warning?
				continue
			}
			if !pe.IsStepComplete(dep) {
				dependendenciesMet = false
				break
			}
		}
		if !dependendenciesMet {
			continue
		}

		// Plan to run the step.
		e.NextSteps = append(e.NextSteps, step.Name)
	}
	if err := h.EventBus.Publish(ctx, &e); err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePlanToPipelineFailed(cmd, err)))
	}

	return nil

}
