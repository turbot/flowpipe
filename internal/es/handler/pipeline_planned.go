package handler

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
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
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePlanned")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
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
		// PRE: No new steps to execute, so the planner should just check to see if
		// all existing steps are complete.
		if pe.IsComplete() {
			if pe.ShouldFail() {
				// There's no error supplied here because it's the step failure that is causing the pipeline to fail
				cmd := event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, nil))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				return h.CommandBus.Send(ctx, &cmd)
			} else {
				cmd, err := event.NewPipelineFinish(event.ForPipelinePlannedToPipelineFinish(e))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				return h.CommandBus.Send(ctx, &cmd)
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
		cmd := event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, perr.InternalWithMessage("pipeline failed")))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, &cmd)
	}

	// PRE: The planner has told us what steps to run next, our job is to start them
	for _, nextStep := range e.NextSteps {
		stepDefn := pipelineDefn.GetStep(nextStep.StepName)

		var evalContext *hcl.EvalContext

		// ! This is a slice of map. Each slice represent a step execution, an element of the for_each
		forEachCtyVals := []map[string]cty.Value{}
		forEachNextStepAction := []modconfig.NextStepAction{}
		stepForEach := stepDefn.GetForEach()

		// if we have for_each build the list of inputs for the for_each
		if stepForEach != nil {
			var err error
			evalContext, err = ex.BuildEvalContext(pipelineDefn, pe)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}

			// First we want to evaluate the content of for_each
			// Given the following:
			//
			// params = ["brian", "freddie"]
			// for_each = params.users
			//
			// The result of "val" should be a CTY List of 2 elements: brian and freddie.
			//
			// Each element in the array represent a "new" step execution. A non-for_each step execution will just have one input
			// so if a step has a for_each we need to build the list if input. Each element in the list represents a step execution.
			val, diags := stepForEach.Value(evalContext)

			if diags.HasErrors() {
				err := error_helpers.HclDiagsToError("param", diags)
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}

			if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
				listVal := val.AsValueSlice()
				for i, v := range listVal {
					forEachCtyVals = append(forEachCtyVals, map[string]cty.Value{
						schema.AttributeTypeValue: v,
						schema.AttributeKey:       cty.NumberIntVal(int64(i)),
					})
				}
			} else if val.Type().IsMapType() || val.Type().IsObjectType() {
				mapVal := val.AsValueMap()
				for k, v := range mapVal {
					forEachCtyVals = append(forEachCtyVals, map[string]cty.Value{
						schema.AttributeTypeValue: v,
						schema.AttributeKey:       cty.StringVal(k),
					})
				}
			}
		}

		// inputs will gather the input data for each step execution, if we have a for_each
		// the inputs length maybe > 1. If we don't have a for_each, then the inputs length will be
		// exactly 1
		//
		inputs := []modconfig.Input{}

		if evalContext == nil {
			var err error
			evalContext, err = ex.BuildEvalContext(pipelineDefn, pe)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
		}

		// now resolve the inputs, if there's no for_each then there's just one input
		if len(forEachCtyVals) == 0 {

			calculateInput := true

			if stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf] != nil {
				expr := stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf]

				val, diags := expr.Value(evalContext)
				if len(diags) > 0 {
					err := error_helpers.HclDiagsToError("diags", diags)

					logger.Error("Error evaluating if condition", "error", err)
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}

				if val.False() {
					logger.Info("if condition not met for step", "step", stepDefn.GetName())
					calculateInput = false
					forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionSkip)
				} else {
					forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionStart)
				}
			} else {
				forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionStart)
			}

			if calculateInput {
				// There's no for_each
				stepInputs, err := stepDefn.GetInputs(evalContext)
				if err != nil {
					logger.Error("Error resolving step inputs for single step", "error", err)
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				inputs = append(inputs, stepInputs)
			} else {
				// If we're to skip the next step, then we need to add a dummy input
				inputs = append(inputs, map[string]interface{}{})
			}
		} else {

			// We have for_each!
			for _, v := range forEachCtyVals {

				// "each" is the magic keyword that will be used to access the current element in the for_each
				//
				// flowpipe's step must use the "each" keyword to access the for_each element that it's currently running
				evalContext.Variables[schema.AttributeEach] = cty.ObjectVal(v)

				// check the "IF" block to see if the step should be skipped?
				// I used to do this in the "step_start" section, but if the IF attribute uses the "each" element, this is the place
				// to do it

				calculateInput := true
				if stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf] != nil {
					expr := stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf]

					val, diags := expr.Value(evalContext)
					if len(diags) > 0 {
						err := error_helpers.HclDiagsToError("diags", diags)
						logger.Error("Error evaluating if condition", "error", err)
						return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					}

					if val.False() {
						logger.Info("if condition not met for step", "step", stepDefn.GetName())
						calculateInput = false
						forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionSkip)
					} else {
						forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionStart)
					}
				} else {
					forEachNextStepAction = append(forEachNextStepAction, modconfig.NextStepActionStart)
				}

				if calculateInput {
					stepInputs, err := stepDefn.GetInputs(evalContext)
					if err != nil {
						logger.Error("Error resolving step inputs for for_each step", "error", err)
						return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					}
					inputs = append(inputs, stepInputs)
				} else {
					// If we're to skip the next step, then we need to add a dummy input
					inputs = append(inputs, map[string]interface{}{})
				}

			}
		}

		// If we have a for_each then the input will be expanded to the number of elements in the for_each
		for i, input := range inputs {

			// Start each step in parallel
			go func(nextStep modconfig.NextStep, input modconfig.Input, index int) {

				var forEachControl *modconfig.StepForEach

				// TODO: this is not very nice and the only reason we do this is for the snapshot, we should
				// TODO: refactor this
				if stepForEach == nil {
					forEachControl = nil
				} else {
					forEachIndex := index
					forEachCtyVal := forEachCtyVals[index][schema.AttributeTypeValue]

					var title string

					if forEachCtyVal.Type().IsPrimitiveType() {
						title += forEachCtyVal.AsString()
					} else {
						title += nextStep.StepName
					}
					forEachOutput := &modconfig.Output{
						Data: map[string]interface{}{},
					}
					forEachOutput.Data[schema.AttributeTypeValue] = title

					forEachControl = &modconfig.StepForEach{
						Index:      forEachIndex,
						Output:     forEachOutput,
						TotalCount: len(inputs),
						Each:       json.SimpleJSONValue{Value: forEachCtyVal},
					}
				}

				cmd, err := event.NewPipelineStepQueue(event.PipelineStepQueueForPipelinePlanned(e), event.PipelineStepQueueWithStep(nextStep.StepName, input, forEachControl, nextStep.DelayMs, forEachNextStepAction[index]))
				if err != nil {
					err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					if err != nil {
						logger.Error("Error publishing event", "error", err)
					}

					return
				}

				if err := h.CommandBus.Send(ctx, &cmd); err != nil {
					err := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					if err != nil {
						logger.Error("Error publishing event", "error", err)
					}
					return
				}
			}(nextStep, input, i)
		}
	}

	return nil
}
