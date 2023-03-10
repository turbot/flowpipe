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
		return err
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		e := event.PipelineFailed{
			Event:        event.NewFlowEvent(cmd.Event),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e, err := event.NewPipelinePlanned(event.ForPipelinePlan(cmd))
	if err != nil {
		return err
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

		// If the step is already planned, then skip it
		if pe.StepStatus[step.Name].Status != "" {
			continue
		}

		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range step.DependsOn {
			stepDefn, ok := defn.Steps[dep]
			if !ok {
				// Dependency is not defined in the pipeline
				// TODO - issue a warning?
				continue
			}
			if stepDefn.Name == dep {
				// Cannot depend on yourself
				// TODO - issue a warning?
				continue
			}
			if pe.StepStatus[dep].Status != "finished" {
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
		return err
	}

	// The planner also needs to check for any child pipelines that have
	// finished and trigger the step finished event.

	/*
		for _, stepExecutionID := range pe.StepExecutions {
			se := ex.StepExecutions[stepExecutionID]
			if se.Status != "started" {
				// Only matters for running child pipelines
				continue
			}
			sd, err := ex.StepDefinition(se.Name)
			if err != nil {
				return err
			}
			if sd.Type != "pipeline" {
				// Only matters for pipeline steps
				continue
			}
			if se.Status == "started" {
				// TODO - this is a bit of a hack. We should be able to
				// trigger the step finished event directly, but it is
				// currently not possible to create a StepFinished event
				// without a PipelineFinished event.
				cmd, err := event.NewPipelineFinished(event.ForChildPipelineFinished(cmd, se.ID))
				if err != nil {
					return err
				}
				return h.EventBus.Send(ctx, &cmd)
			}
		}
	*/

	return nil

}
