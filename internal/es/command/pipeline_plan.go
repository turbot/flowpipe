package command

import (
	"context"
	"log/slog"
	"sync"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return execution.PipelinePlanCommand.HandlerName()
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &event.PipelinePlan{}
}

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd, ok := c.(*event.PipelinePlan)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelinePlan", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelinePlan")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(cmd.Event.ExecutionID, cmd.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_plan: Error loading pipeline execution", "error", err)
		return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, "", "")
	}
	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	// Create a new PipelinePlanned event
	e, err := event.NewPipelinePlanned(event.ForPipelinePlan(cmd))
	if err != nil {
		return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, "")
	}

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
	if err != nil {
		slog.Error("Error building eval context for step", "error", err)
		return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, "")
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
	for _, stepDefn := range pipelineDefn.Steps {

		// This mean the step has been initialized
		if pex.StepStatus[stepDefn.GetFullyQualifiedName()] != nil {
			continue
		}

		if pex.IsStepQueued(stepDefn.GetFullyQualifiedName()) {
			continue
		}

		// No need to plan if the step has been initialized
		if pex.IsStepInitialized(stepDefn.GetFullyQualifiedName()) {
			continue
		}

		// If the steps dependencies are not met, then skip it.
		// TODO - this is completely naive and does not handle cycles.
		dependendenciesMet := true
		for _, dep := range stepDefn.GetDependsOn() {

			// Cannot depend on yourself
			if stepDefn.GetFullyQualifiedName() == dep {
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

			// Do not check for ignore error = true here. It may have been overriden by the "Failure Mode = evaluation" directive. The right place
			// to do this is in the execution layer where we build the "step status"
			if pex.IsStepFail(dep) {
				dependendenciesMet = false

				// If one of the dependencies failed, and it is not ignored, AND it is the final failure, then this
				// step will never start. Put it down in the "Inaccessible" list so we know that the Pipeline must
				// be ended in the handler/pipeline_planned stage
				e.NextSteps = append(e.NextSteps, flowpipe.NextStep{
					StepName: stepDefn.GetFullyQualifiedName(),
					Action:   flowpipe.NextStepActionInaccessible})
				break
			}
		}

		if !dependendenciesMet {
			continue
		}

		maxConcurency := stepDefn.GetMaxConcurrency(evalContext)

		nextStep := flowpipe.NextStep{
			StepName:       stepDefn.GetFullyQualifiedName(),
			Action:         flowpipe.NextStepActionStart,
			MaxConcurrency: maxConcurency,
		}

		// Check if there's a for_each.
		//
		// If for_each does not exist: calculate the input
		// if for_each exists:  don't calculate the input, it's the job of step_for_each_plan to calculate the input
		//
		stepForEach := stepDefn.GetForEach()
		if helpers.IsNil(stepForEach) {

			// there's no step for each
			var nextStepAction flowpipe.NextStepAction
			var input flowpipe.Input

			if !helpers.IsNil(stepDefn.GetLoopConfig()) {
				// If the execution falls here, it means it's the beginning of the loop
				// if it's part of a loop, it will be short circuited in the beginning of this for loop
				evalContext = execution.AddLoop(nil, evalContext)
			}

			calculateInput := true

			// Check if the step needs to run or skip (that's the IF block)
			if stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf] != nil {
				expr := stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf]

				val, diags := expr.Value(evalContext)
				if len(diags) > 0 {
					err := error_helpers.HclDiagsToError("diags", diags)

					slog.Error("Error evaluating if condition", "error", err)
					return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
				}

				if val.False() {
					slog.Debug("if condition not met for step", "step", stepDefn.GetName())
					calculateInput = false
					nextStepAction = flowpipe.NextStepActionSkip
				} else {
					nextStepAction = flowpipe.NextStepActionStart
				}
			} else {
				nextStepAction = flowpipe.NextStepActionStart
			}

			if calculateInput {
				// There's no for_each
				evalContext, err = ex.AddCredentialsToEvalContext(evalContext, stepDefn)
				if err != nil {
					slog.Error("Error adding credentials to eval context", "error", err)
					return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
				}

				evalContext, err := ex.AddConnectionsToEvalContextWithForEach(evalContext, stepDefn, pipelineDefn, false, nil)
				if err != nil {
					slog.Error("Error adding connections to eval context during pipeline plan (1)", "error", err)
					return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
				}

				stepInputs, connDepend, err := stepDefn.GetInputs2(evalContext)
				if err != nil {
					return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
				}

				if len(connDepend) > 0 {
					evalContext, err := ex.AddConnectionsToEvalContextWithForEach(evalContext, stepDefn, pipelineDefn, false, connDepend)
					if err != nil {
						slog.Error("Error adding connections to eval context during pipeline plan (2)", "error", err)
						return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
					}
					var connDepend2 []flowpipe.ConnectionDependency
					stepInputs, connDepend2, err = stepDefn.GetInputs2(evalContext)
					if err != nil {
						return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, err, pex.Name, stepDefn.GetName())
					}
					if len(connDepend2) > 0 {
						// we are missing some connections
						missingConnStr := ""
						for i, connDep := range connDepend2 {
							if i > 0 {
								missingConnStr += ", "
							}
							missingConnStr += connDep.Type
							if connDep.Source != "" {
								missingConnStr += "." + connDep.Source
							}
						}
						missingConnErrors := perr.InternalWithMessage("Missing connections for step '" + stepDefn.GetName() + "': " + missingConnStr)
						slog.Error("Missing connections for step", "step", stepDefn.GetName(), "missing", missingConnStr)
						return h.raiseNewPipelineFailedEvent(ctx, plannerMutex, cmd, missingConnErrors, pex.Name, stepDefn.GetName())
					}
				}

				// There's no for_each, there's only a single input
				input = stepInputs
			} else {
				// If we're to skip the next step, then we need to add a dummy input
				input = map[string]interface{}{}
			}

			nextStep.Input = input
			nextStep.Action = nextStepAction
		}

		// Plan to run the step.
		e.NextSteps = append(e.NextSteps, nextStep)
	}

	// Pipeline has been planned, now publish this event
	if err := h.EventBus.Publish(ctx, e); err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(cmd, err, pex.Name, "")))
	}

	return nil
}

func (h PipelinePlanHandler) raiseNewPipelineFailedEvent(ctx context.Context, plannerMutex *sync.Mutex, cmd *event.PipelinePlan, err error, pipelineName, stepName string) error {
	publishErr := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePlanToPipelineFailed(cmd, err, pipelineName, stepName)))
	if publishErr != nil {
		slog.Error("pipeline_plan: Error publishing pipeline failed event", "error", publishErr)
	}
	return nil
}
