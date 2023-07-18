package handler

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/zclconf/go-cty/cty"
)

type PipelinePlanned EventHandler

func (h PipelinePlanned) HandlerName() string {
	return "handler.pipeline_planned"
}

func (PipelinePlanned) NewEvent() interface{} {
	return &event.PipelinePlanned{}
}

func (h PipelinePlanned) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelinePlanned)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelinePlanned")
	}

	logger.Info("[9] pipeline planned event handler #1", "executionID", e.Event.ExecutionID)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	defn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[e.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pe.IsCanceled() || pe.IsPaused() || pe.IsFinishing() || pe.IsFinished() {
		return nil
	}

	if len(e.NextSteps) == 0 {
		logger.Info("[9] pipeline planned event handler #2", "executionID", e.Event.ExecutionID)

		// PRE: No new steps to execute, so the planner should just check to see if
		// all existing steps are complete.
		if pe.IsComplete() {
			if pe.ShouldFail() {
				logger.Info("[9] pipeline planned event handler #4 - should fail", "executionID", e.Event.ExecutionID)

				// TODO: what is the error on the pipeline?
				cmd := event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, fperr.InternalWithMessage("pipeline failed")))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				return h.CommandBus.Send(ctx, &cmd)
			} else {

				logger.Info("[9] pipeline planned event handler #5 - complete", "executionID", e.Event.ExecutionID)
				cmd, err := event.NewPipelineFinish(event.ForPipelinePlannedToPipelineFinish(e))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				return h.CommandBus.Send(ctx, &cmd)
			}
		}

		return nil
	}

	// PRE: The planner has told us what steps to run next, our job is to start them
	for _, nextStep := range e.NextSteps {

		// data, err := ex.PipelineData(e.PipelineExecutionID)
		_, err := ex.PipelineData(e.PipelineExecutionID)
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
		}

		stepDefn := defn.GetStep(nextStep.StepName)

		evalContext := hcl.EvalContext{
			Variables: ex.ExecutionVariables,
			Functions: pipeparser.ContextFunctions(viper.GetString("work.dir")),
		}

		// ! This is a slice of map. Each slice represent a step execution, an element of the for_each
		forEachCtyVals := []map[string]cty.Value{}
		stepForEach := stepDefn.GetForEach()
		if stepForEach != nil {
			// TODO: this only works for default params - we need to implement the parameter passing when
			// TODO: running the pipeline
			paramsCtyVal, err := defn.ParamsAsCty()
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
			evalContext.Variables["param"] = paramsCtyVal

			val, diags := stepForEach.Value(&evalContext)

			if diags.HasErrors() {
				err := pipeparser.DiagsToError("param", diags)
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}

			if val.Type().IsListType() {
				listVal := val.AsValueSlice()
				for _, v := range listVal {
					forEachCtyVals = append(forEachCtyVals, map[string]cty.Value{
						"value": v,
					})
				}
			} else {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, fperr.InternalWithMessage("for_each must be a list (for now)"))))
			}
		}

		// inputs will gather the input data for each step execution, if we have a for_each
		// the inputs length maybe > 1. If we don't have a for_each, then the inputs length will be
		// exactly 1
		//
		inputs := []types.Input{}

		// now resolve the inputs, if there's no for_each then there's just one input
		if len(forEachCtyVals) == 0 {
			stepInputs, err := stepDefn.GetInputs(&evalContext)
			if err != nil {
				logger.Error("Error resolving step inputs", "error", err)
				return err
			}
			inputs = append(inputs, stepInputs)
		} else {
			for _, v := range forEachCtyVals {
				evalContext.Variables["each"] = cty.ObjectVal(v)
				stepInputs, err := stepDefn.GetInputs(&evalContext)
				if err != nil {
					logger.Error("Error resolving step inputs", "error", err)
					return err
				}
				inputs = append(inputs, stepInputs)
			}
		}

		for i, input := range inputs {
			// Start each step in parallel
			go func(nextStep types.NextStep, input types.Input, index int) {
				var forEachIndex *int
				var forEachOutput *types.StepOutput

				// TODO: this is not very nice and the only reason we do this is for the snapshot, we should
				// TODO: refactor this
				if stepForEach == nil {
					forEachIndex = nil
					forEachOutput = nil
				} else {
					forEachIndex = &index
					forEachCtyVal := forEachCtyVals[index]["value"]

					var title string

					if forEachCtyVal.Type().IsPrimitiveType() {
						title += forEachCtyVal.AsString()
					} else {
						title += nextStep.StepName
					}

					forEachOutput = &types.StepOutput{
						"value": title,
					}
				}

				cmd, err := event.NewPipelineStepQueue(event.PipelineStepQueueForPipelinePlanned(e), event.PipelineStepQueueWithStep(nextStep.StepName, input, forEachIndex, forEachOutput, nextStep.DelayMs))
				if err != nil {
					err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					if err != nil {
						fplog.Logger(ctx).Error("Error publishing event", "error", err)
					}

					return
				}

				logger.Info("[8] pipeline planned event handler #3.B - sending pipeline step QUEUE command", "command", cmd)
				if err := h.CommandBus.Send(ctx, &cmd); err != nil {
					err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					if err != nil {
						fplog.Logger(ctx).Error("Error publishing event", "error", err)
					}
					return
				}
			}(nextStep, input, i)
		}
	}

	return nil
}
