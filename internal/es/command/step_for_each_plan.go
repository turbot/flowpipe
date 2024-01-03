package command

import (
	"context"
	"strconv"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type StepForEachPlanHandler CommandHandler

var stepForEachPlan = event.StepForEachPlan{}

func (h StepForEachPlanHandler) HandlerName() string {
	return stepForEachPlan.HandlerName()
}

func (h StepForEachPlanHandler) NewCommand() interface{} {
	return &event.StepForEachPlan{}
}

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

func (h StepForEachPlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd, ok := c.(*event.StepForEachPlan)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.StepForEachPlan", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.StepForEachPlan")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(cmd.Event.ExecutionID, cmd.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_plan: Error loading pipeline execution", "error", err)
		return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
	}
	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	stepDefn := pipelineDefn.GetStep(cmd.StepName)
	if stepDefn == nil {
		slog.Error("step not found", "step_name", cmd.StepName)
		return h.raiseNewPipelineFailedEvent(ctx, cmd, perr.BadRequestWithMessage("step not found"))
	}

	stepForEach := stepDefn.GetForEach()
	if helpers.IsNil(stepForEach) {
		slog.Error("step does not have a for_each", "step_name", cmd.StepName)
		return h.raiseNewPipelineFailedEvent(ctx, cmd, perr.BadRequestWithMessage("step does not have a for_each"))
	}

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
	if err != nil {
		slog.Error("Error building eval context for for_each", "error", err)
		return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
	}

	if stepDefn.GetUnresolvedBodies()["loop"] != nil {
		// If the execution falls here, it means it's the beginning of the loop
		// if it's part of a loop, it will be short circuited in the beginning of this for loop
		evalContext = execution.AddLoop(nil, evalContext)
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
		return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
	}

	forEachCtyVals := map[string]map[string]cty.Value{}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		listVal := val.AsValueSlice()
		for i, v := range listVal {
			forEachCtyVals[strconv.Itoa(i)] = map[string]cty.Value{
				schema.AttributeTypeValue: v,
				schema.AttributeKey:       cty.NumberIntVal(int64(i)),
			}
		}
	} else if val.Type().IsMapType() || val.Type().IsObjectType() {
		mapVal := val.AsValueMap()
		for k, v := range mapVal {
			forEachCtyVals[k] = map[string]cty.Value{
				schema.AttributeTypeValue: v,
				schema.AttributeKey:       cty.StringVal(k),
			}
		}
	} else {
		err := perr.BadRequestWithMessage("for_each must be a list, set, tuple, map or object for step " + stepDefn.GetName())
		return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
	}

	var nextSteps []modconfig.NextStep

	stepStatusList := pex.StepStatus[cmd.StepName]

	// forEachCtyVals is a map the key is either a string of "0", "1", "2" (string! not index of a slice)
	// or a string of a key in a map when the for_each is against a map attributes:
	//  {
	//     foo: bar
	//     baz: quz
	//   }
	//
	//  in the above map the key are foo and baz
	for k, v := range forEachCtyVals {

		nextStep := modconfig.NextStep{
			StepName: cmd.StepName,
		}

		// check the current execution if the step is already completed (or failed)
		stepStatus := stepStatusList[k]
		if stepStatus != nil {
			if stepStatus.IsComplete() {
				continue
			}

			if stepStatus.Initializing || len(stepStatus.Queued) > 0 || len(stepStatus.Started) > 0 || len(stepStatus.Finished) > 0 || len(stepStatus.Failed) > 0 ||
				stepStatus.ErrorHold || stepStatus.LoopHold {
				continue
			}
		}

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
				slog.Error("Error evaluating if condition", "error", err)
				return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
			}

			if val.False() {
				slog.Debug("if condition not met for step", "step", stepDefn.GetName())
				calculateInput = false
				nextStep.Action = modconfig.NextStepActionSkip
			} else {
				nextStep.Action = modconfig.NextStepActionStart
			}
		} else {
			nextStep.Action = modconfig.NextStepActionStart
		}

		if calculateInput {
			evalContext, err = ex.AddCredentialsToEvalContext(evalContext, stepDefn)
			if err != nil {
				slog.Error("Error adding credentials to eval context", "error", err)
				return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
			}

			stepInputs, err := stepDefn.GetInputs(evalContext)
			if err != nil {
				slog.Error("Error resolving step inputs for for_each step", "error", err)
				return h.raiseNewPipelineFailedEvent(ctx, cmd, err)
			}
			nextStep.Input = stepInputs
		} else {
			// If we're to skip the next step, then we need to add a dummy input
			nextStep.Input = map[string]interface{}{}
		}

		forEachCtyVal := forEachCtyVals[k][schema.AttributeTypeValue]
		forEachControl := &modconfig.StepForEach{
			ForEachStep: true,
			Key:         k,
			TotalCount:  len(forEachCtyVals),
			Each:        json.SimpleJSONValue{Value: forEachCtyVal},
		}
		nextStep.StepForEach = forEachControl

		nextSteps = append(nextSteps, nextStep)
	}

	err = h.EventBus.Publish(ctx, event.NewStepForEachPlannedFromStepForEachPlan(cmd, nextSteps))
	if err != nil {
		err = h.raiseNewPipelineFailedEvent(ctx, cmd, err)
		if err != nil {
			slog.Error("Error publishing new pipeline failed event", "error", err)
		}
	}
	return nil
}

func (h StepForEachPlanHandler) raiseNewPipelineFailedEvent(ctx context.Context, e *event.StepForEachPlan, err error) error {
	publishErr := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepForEachPlan(e, err))
	if publishErr != nil {
		slog.Error("Error publishing pipeline failed event", "error", publishErr)
	}
	return nil
}
