package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type PipelinePlanned EventHandler

var pipelinePlanned = event.PipelinePlanned{}

func (h PipelinePlanned) HandlerName() string {
	return pipelinePlanned.HandlerName()
}

func (PipelinePlanned) NewEvent() interface{} {
	return &event.PipelinePlanned{}
}

func (h PipelinePlanned) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelinePlanned)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePlanned")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
	}

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
	}

	// Convenience
	pe := ex.PipelineExecutions[e.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pe.IsCanceled() || pe.IsPaused() || pe.IsFinishing() || pe.IsFinished() {
		return nil
	}

	if len(e.NextSteps) == 0 {
		// PRE: No new steps to execute, so the planner should just check to see if
		// all existing steps are complete.
		if pe.IsComplete() {
			if pe.ShouldFail() {
				// There's no error supplied here because it's the step failure that is causing the pipeline to fail
				cmd := event.NewPipelineFailFromPipelinePlanned(e, nil)
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
				}
				return h.CommandBus.Send(ctx, cmd)
			} else {
				cmd, err := event.NewPipelineFinish(event.ForPipelinePlannedToPipelineFinish(e))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
				}
				return h.CommandBus.Send(ctx, cmd)
			}
		}

		return nil
	}

	// Check if there is a step that is "inaccessible", if so then we terminate the pipeline
	// since there's no possibility of it ever completing
	// TODO: there's optimisation here, we could potentially run all the other steps that can run
	// TODO: but for now take the simplest route
	pipelineInaccessible := false
	for _, nextStep := range e.NextSteps {
		if nextStep.Action == modconfig.NextStepActionInaccessible {
			pipelineInaccessible = true
			break
		}
	}

	if pipelineInaccessible {
		logger.Info("Pipeline is inaccessible, terminating", "pipeline", pipelineDefn.Name)
		// TODO: what is the error on the pipeline?
		cmd := event.NewPipelineFailFromPipelinePlanned(e, perr.InternalWithMessage("pipeline failed"))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// PRE: The planner has told us what steps to run next, our job is to start them
	for _, nextStep := range e.NextSteps {

		stepDefn := pipelineDefn.GetStep(nextStep.StepName)

		if nextStep.StepLoop != nil {
			panic("loop not supported")

			// Special instruction for "loop"
			// hasForEach := false
			// inputCount := 1
			// key := "0"

			// input := *nextStep.StepLoop.Input

			// var foreachOutput modconfig.Output

			// // calculate the for_each control, is it a single step or a for_each step?
			// forEachCtyVal := cty.StringVal("0")
			// if nextStep.StepForEach != nil {
			// 	hasForEach = true
			// 	inputCount = nextStep.StepForEach.TotalCount
			// 	key = nextStep.StepForEach.Key
			// 	eachGoVal, err := hclhelpers.CtyToGo(nextStep.StepForEach.Each.Value)
			// 	if err != nil {
			// 		logger.Error("Error converting cty to go", "error", err)
			// 		return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
			// 	}
			// 	input[schema.AttributeEach] = eachGoVal
			// 	// foreachOutput = *nextStep.StepForEach.Output
			// 	forEachCtyVal = nextStep.StepForEach.Each.Value
			// }

			// // in a "loop" we should already know the input, it's a side effect of calculating the "IF" attribute of the loop
			// go runNonForEachStep(ctx, h.CommandBus, e, hasForEach, foreachOutput, forEachCtyVal, inputCount, modconfig.NextStepActionStart, nextStep, input, key)

			// continue
		}

		stepForEach := stepDefn.GetForEach()

		// If there's a for each, runs a new command: step_for_each_planner
		if !helpers.IsNil(stepForEach) {
			var err error

			stepForEachPlanCmd := event.NewStepForEachPlanFromPipelinePlanned(e, nextStep.StepName)
			err = h.CommandBus.Send(ctx, stepForEachPlanCmd)

			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
			}

			// don't return here, but process the next step
			continue
		}

		forEachNextStepActions := map[string]modconfig.NextStepAction{}

		// inputs will gather the input data for each step execution, if we have a for_each
		// the inputs length maybe > 1. If we don't have a for_each, then the inputs length will be
		// exactly 1
		//
		input := nextStep.Input

		// evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
		// if err != nil {
		// 	logger.Error("Error building eval context for step", "error", err)
		// 	return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		// }
		// if stepDefn.GetUnresolvedBodies()["loop"] != nil {
		// 	// If the execution falls here, it means it's the beginning of the loop
		// 	// if it's part of a loop, it will be short circuited in the beginning of this for loop
		// 	// evalContext = execution.AddLoop(nil, evalContext)
		// }

		// var title string

		// if forEachCtyVal.Type().IsPrimitiveType() {
		// 	t, err := hclhelpers.CtyToString(forEachCtyVal)
		// 	if err != nil {
		// 		logger.Error("Error converting cty to string", "error", err)
		// 	} else {
		// 		title += t
		// 	}
		// } else {
		// 	title += nextStep.StepName
		// }
		// forEachOutput := modconfig.Output{
		// 	Data: map[string]interface{}{},
		// }
		// forEachOutput.Data[schema.AttributeTypeValue] = title

		// Start each step in parallel
		runNonForEachStep(ctx, h.CommandBus, e, modconfig.Output{}, forEachNextStepActions["0"], nextStep, input, "0")

	}

	return nil
}

func runNonForEachStep(ctx context.Context, commandBus *FpCommandBus, e *event.PipelinePlanned, forEachOutput modconfig.Output, forEachNextStepAction modconfig.NextStepAction, nextStep modconfig.NextStep, input modconfig.Input, key string) {

	logger := fplog.Logger(ctx)

	// If a step does not have a for_each, we still build a for_each control but with key of "0"
	forEachControl := &modconfig.StepForEach{
		Key:        "0",
		TotalCount: 1,
		Each:       json.SimpleJSONValue{Value: cty.StringVal("0")},
	}

	cmd, err := event.NewPipelineStepQueue(event.PipelineStepQueueForPipelinePlanned(e), event.PipelineStepQueueWithStep(nextStep.StepName, input, forEachControl, nextStep.StepLoop, nextStep.DelayMs, forEachNextStepAction))
	if err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}

		return
	}

	if err := commandBus.Send(ctx, cmd); err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}
		return
	}
}
