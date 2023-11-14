package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type PipelinePlanned EventHandler

func (h PipelinePlanned) HandlerName() string {
	return execution.PipelinePlannedEvent.HandlerName()
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

		loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
		var stepLoop *modconfig.StepLoop
		if loopBlock != nil {
			stepLoop = &modconfig.StepLoop{
				Index: 0,
			}
		}

		// Start each step in parallel
		runNonForEachStep(ctx, h.CommandBus, e, modconfig.Output{}, nextStep.Action, nextStep, nextStep.Input, stepLoop)
	}

	return nil
}

func runNonForEachStep(ctx context.Context, commandBus *FpCommandBus, e *event.PipelinePlanned, forEachOutput modconfig.Output, forEachNextStepAction modconfig.NextStepAction, nextStep modconfig.NextStep, input modconfig.Input, stepLoop *modconfig.StepLoop) {

	logger := fplog.Logger(ctx)

	// If a step does not have a for_each, we still build a for_each control but with key of "0"
	forEachControl := &modconfig.StepForEach{
		ForEachStep: false,
		Key:         "0",
		TotalCount:  1,
		Each:        json.SimpleJSONValue{Value: cty.StringVal("0")},
	}

	cmd, err := event.NewStepQueue(event.StepQueueForPipelinePlanned(e), event.StepQueueWithStep(nextStep.StepName, input, forEachControl, nextStep.StepLoop, forEachNextStepAction))
	cmd.StepLoop = stepLoop

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
