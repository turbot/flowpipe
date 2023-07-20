package command

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return "command.pipeline_plan"
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &event.PipelinePlan{}
}

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {

	logger := fplog.Logger(ctx)

	evt, ok := c.(*event.PipelinePlan)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelinePlan", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelinePlan")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		logger.Error("pipeline_plan: Error loading pipeline execution", "error", err)
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[evt.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pe.IsCanceled() || pe.IsPaused() || pe.IsFinishing() || pe.IsFinished() {
		return nil
	}

	defn, err := ex.PipelineDefinition(evt.PipelineExecutionID)
	if err != nil {
		logger.Error("Error loading pipeline definition", "error", err)
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

	// Create a new PipelinePlanned event
	e, err := event.NewPipelinePlanned(event.ForPipelinePlan(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

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
	//

	for i, step := range defn.Steps {
		// TODO: this entire failure handling doesn't work since we've moved to HCL.
		if pe.IsStepFail(step.GetFullyQualifiedName()) {

			if !pe.IsStepFinalFailure(defn.Steps[i], ex) {

				// TODO: this won't work with multiple executions of the same step (if we have a FOR step)
				if !pe.IsStepQueued(step.GetFullyQualifiedName()) {
					e.NextSteps = append(e.NextSteps, types.NextStep{StepName: step.GetFullyQualifiedName(), DelayMs: 3000})
				}
			}
			continue
		}

		if pe.IsStepQueued(step.GetFullyQualifiedName()) {
			continue
		}

		// No need to plan if the step has been initialized
		if pe.IsStepInitialized(step.GetFullyQualifiedName()) {
			continue
		}

		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range step.GetDependsOn() {

			// Cannot depend on yourself
			if step.GetFullyQualifiedName() == dep {
				// TODO - issue a warning? How do we issue a warning?
				continue
			}
			// Ignore invalid dependencies
			if defn.GetStep(dep) == nil {
				// TODO - issue a warning? How do we issue a warning?
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
		e.NextSteps = append(e.NextSteps, types.NextStep{StepName: step.GetFullyQualifiedName()})
	}

	// Pipeline has been planned, now publish this event
	if err := h.EventBus.Publish(ctx, &e); err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

	return nil

}
