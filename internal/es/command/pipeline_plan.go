package command

import (
	"context"
	"sync"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return "command.pipeline_plan"
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &event.PipelinePlan{}
}

// Define a mutex.
var lock sync.Mutex

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	evt, ok := c.(*event.PipelinePlan)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelinePlan", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelinePlan")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		logger.Error("pipeline_plan: Error loading pipeline execution", "error", err)
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

	// Convenience
	pex := ex.PipelineExecutions[evt.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	pipelineDefn, err := ex.PipelineDefinition(evt.PipelineExecutionID)
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
	for _, step := range pipelineDefn.Steps {
		// TODO: error handling

		// means step has a for_each, each for_each is another "series" of steps
		//
		// the planner need to handle them as if they are invidual "steps"
		//
		// if there's a problem if one of the n number of for_each, we just want to retry that one
		//
		// for example
		/*
			   step "echo" "echo {
					for_each = ["foo", "bar"]
					text = "foo"
			   }

			   this step will generate 2 "index".
		*/

		// This mean the step has been initialized

		if len(pex.StepStatus[step.GetFullyQualifiedName()]) > 0 {

			// for_each that returns a list will still be a map, but the key of the map is a string
			// of "0", "1", "2" and so on.
			for _, stepStatus := range pex.StepStatus[step.GetFullyQualifiedName()] {

				if stepStatus.StepExecutions == nil {
					continue
				}

				// find the latest step execution, check if it has a loop that needs to be run
				latestStepExecution := stepStatus.StepExecutions[len(stepStatus.StepExecutions)-1]

				// TODO: error retry

				// no step loop means we're done here
				if latestStepExecution.StepLoop == nil || latestStepExecution.StepLoop.LoopCompleted {
					continue
				}

				// Just because the loop has not been completed, it doesn't mean the next step is NOT already been started by another planner (!)
				// check the queue status
				//
				// TODO: locking issue
				if len(stepStatus.Queued) > 0 || len(stepStatus.Started) > 0 {
					continue
				}

				// bypass depends_on check because if we're here, the step has already started so we know that all its
				// dependencies are met
				//
				e.NextSteps = append(e.NextSteps, modconfig.NextStep{
					StepName:    step.GetFullyQualifiedName(),
					Action:      modconfig.NextStepActionStart,
					StepForEach: latestStepExecution.StepForEach,
					StepLoop:    latestStepExecution.StepLoop,
				})
			}

			continue
		}

		if pex.IsStepQueued(step.GetFullyQualifiedName()) {
			continue
		}

		// No need to plan if the step has been initialized
		if pex.IsStepInitialized(step.GetFullyQualifiedName()) {
			continue
		}

		if pex.IsStepInLoopHold(step.GetFullyQualifiedName()) {
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
			depStepDefn := pipelineDefn.GetStep(dep)
			if depStepDefn == nil {
				// TODO - issue a warning? How do we issue a warning?
				continue
			}

			if !pex.IsStepComplete(dep) {
				dependendenciesMet = false
				break
			}

			if pex.IsStepFail(dep) && (depStepDefn.GetErrorConfig() == nil || !depStepDefn.GetErrorConfig().Ignore) {
				dependendenciesMet = false

				// TODO: final failure is always TRUE for now
				if pex.IsStepFinalFailure(depStepDefn, ex) {
					// If one of the dependencies failed, and it is not ignored, AND it is the final failure, then this
					// step will never start. Put it down in the "Inaccessible" list so we know that the Pipeline must
					// be ended in the handler/pipeline_planned stage
					e.NextSteps = append(e.NextSteps, modconfig.NextStep{
						StepName: step.GetFullyQualifiedName(),
						Action:   modconfig.NextStepActionInaccessible})
				}
				break
			}

		}

		if !dependendenciesMet {
			continue
		}

		// Plan to run the step.
		e.NextSteps = append(e.NextSteps, modconfig.NextStep{
			StepName: step.GetFullyQualifiedName(),
			Action:   modconfig.NextStepActionStart})
	}

	// Pipeline has been planned, now publish this event
	if err := h.EventBus.Publish(ctx, &e); err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(evt, err)))
	}

	return nil
}
