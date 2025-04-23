package command

import (
	"context"
	"log/slog"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type StepStartHandler CommandHandler

func (h StepStartHandler) HandlerName() string {
	return execution.StepStartCommand.HandlerName()
}

func (h StepStartHandler) NewCommand() interface{} {
	return &event.StepStart{}
}

// * This is the handler that will actually execute the primitive
// *
// * At the end of the execution it will raise the appropriate event: StepFinished or PipelineFailed
// *
// * Also note the "special" step handler for launching child pipelines
func (h StepStartHandler) Handle(ctx context.Context, c interface{}) error {

	go func(ctx context.Context, c interface{}, h StepStartHandler) {

		cmd, ok := c.(*event.StepStart)
		if !ok {
			slog.Error("invalid command type", "expected", "*event.StepStart", "actual", c)
			return
		}

		plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
		plannerMutex.Lock()
		defer func() {
			if plannerMutex != nil {
				plannerMutex.Unlock()
			}

			execution.ReleaseStepTypeSemaphore(cmd.StepType)
		}()

		executionID := cmd.Event.ExecutionID

		ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(executionID, cmd.PipelineExecutionID)
		if err != nil {
			slog.Error("pipeline_plan: Error loading pipeline execution", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
			return
		}

		stepDefn := pipelineDefn.GetStep(cmd.StepName)

		defer func() {
			if stepDefn.GetType() == schema.BlockTypePipelineStepInput && o.IsServerMode {
				slog.Debug("Step execution is an input step, not releasing semaphore", "step_name", cmd.StepName, "pipeline_execution_id", cmd.PipelineExecutionID)
				return
			} else if stepDefn.GetType() == schema.BlockTypePipelineStepPipeline && cmd.NextStepAction != resources.NextStepActionSkip {
				slog.Debug("Step execution is a pipeline step, not releasing semaphore", "step_name", cmd.StepName, "pipeline_execution_id", cmd.PipelineExecutionID)
				return
			}

			err := execution.ReleasePipelineExecutionStepSemaphore(cmd.PipelineExecutionID, stepDefn)
			if err != nil {
				slog.Error("Error releasing pipeline execution step semaphore", "error", err)
			}
		}()

		stepOutput := make(map[string]interface{})

		pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

		evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
		if err != nil {
			slog.Error("Error building eval context (step start handler)", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
			return
		}

		evalContext, err = ex.AddCredentialsToEvalContext(evalContext, stepDefn)
		if err != nil {
			slog.Error("Error adding credentials to eval context", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
			return
		}

		evalContext, err = ex.AddConnectionsToEvalContextWithForEach(evalContext, stepDefn, pipelineDefn, false, nil)
		if err != nil {
			slog.Error("Error adding connections to eval context during step start", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
			return
		}

		// Check if the step should be skipped. This is determined by the evaluation of the IF clause during the
		// pipeline_plan phase
		if cmd.NextStepAction == resources.NextStepActionSkip {
			output := &resources.Output{
				Status: "skipped",
			}

			endStep(ex, cmd, output, stepOutput, stepDefn, evalContext, ctx, h.EventBus)
			return
		}

		var output *resources.Output

		if o.IsServerMode {
			var feKey *string
			var li, ri *int
			if cmd.StepForEach != nil && cmd.StepForEach.ForEachStep {
				feKey = &cmd.StepForEach.Key
			}
			if cmd.StepLoop != nil {
				li = &cmd.StepLoop.Index
			}
			if cmd.StepRetry != nil {
				i := cmd.StepRetry.Count + 1
				ri = &i
			}
			sp := types.NewServerOutputPrefixWithExecId(cmd.Event.CreatedAt, "pipeline", &cmd.Event.ExecutionID)
			prefix := types.NewParsedEventPrefix(pipelineDefn.PipelineName, &cmd.StepName, feKey, li, ri, &sp)
			pe := types.NewParsedEvent(prefix, cmd.Event.ExecutionID, h.HandlerName(), stepDefn.GetType(), "")
			o.RenderServerOutput(ctx, types.NewParsedEventWithInput(pe, cmd.StepInput, false))
		}

		// Release the lock so we can have multiple steps running at the same time
		plannerMutex.Unlock()
		plannerMutex = nil

		var primitiveError error
		switch stepDefn.GetType() {
		case schema.BlockTypePipelineStepHttp:
			p := primitive.HTTPRequest{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepPipeline:
			p := primitive.RunPipeline{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepEmail:
			p := primitive.Email{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepQuery:
			p := primitive.Query{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepSleep:
			p := primitive.Sleep{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepTransform:
			p := primitive.Transform{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepFunction:
			p := primitive.Function{
				ModPath: pipelineDefn.GetMod().ModPath,
			}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepContainer:
			p := primitive.Container{FullyQualifiedStepName: stepDefn.GetFullyQualifiedName()}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepInput:
			if routerUrl, routed := primitive.GetInputRouter(); routed {
				endStepFunc := func(stepExecution *execution.StepExecution, out *resources.Output) error {
					return EndStepFromApi(ex, stepExecution, pipelineDefn, stepDefn, out, h.EventBus)
				}
				p := primitive.NewRoutedInput(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, pipelineDefn.PipelineName, cmd.StepName, schema.BlockTypePipelineStepInput, routerUrl, endStepFunc)
				cmd.StepInput["router_url"] = routerUrl
				output, primitiveError = p.Run(ctx, cmd.StepInput)
			} else {
				p := primitive.NewInputPrimitive(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, pipelineDefn.PipelineName, cmd.StepName)
				output, primitiveError = p.Run(ctx, cmd.StepInput)
			}
		case schema.BlockTypePipelineStepMessage:
			if routerUrl, routed := primitive.GetInputRouter(); routed {
				endStepFunc := func(stepExecution *execution.StepExecution, out *resources.Output) error {
					return EndStepFromApi(ex, stepExecution, pipelineDefn, stepDefn, out, h.EventBus)
				}
				p := primitive.NewRoutedInput(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, pipelineDefn.PipelineName, cmd.StepName, schema.BlockTypePipelineStepMessage, routerUrl, endStepFunc)
				cmd.StepInput["router_url"] = routerUrl
				output, primitiveError = p.Run(ctx, cmd.StepInput)
			} else {
				p := primitive.NewMessagePrimitive(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, pipelineDefn.PipelineName, cmd.StepName)
				output, primitiveError = p.Run(ctx, cmd.StepInput)
			}
		default:
			slog.Error("Unknown step type", "type", stepDefn.GetType())

			plannerMutex = event.GetEventStoreMutex(cmd.Event.ExecutionID)
			plannerMutex.Lock()

			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}

			return
		}

		plannerMutex = event.GetEventStoreMutex(cmd.Event.ExecutionID)
		plannerMutex.Lock()

		if primitiveError != nil {
			slog.Error("primitive failed", "error", primitiveError)
			if output == nil {
				output = &resources.Output{}
			}
			if output.Errors == nil {
				output.Errors = []resources.StepError{}
			}

			output.Errors = append(output.Errors, resources.StepError{
				Error: perr.InternalWithMessage(primitiveError.Error()),
			})

		}

		// Decorate the errors
		if output.HasErrors() {
			output.Status = constants.StateFailed
			for i := 0; i < len(output.Errors); i++ {
				(output.Errors)[i].Step = cmd.StepName
				(output.Errors)[i].PipelineExecutionID = cmd.PipelineExecutionID
				(output.Errors)[i].StepExecutionID = cmd.StepExecutionID
				(output.Errors)[i].Pipeline = pipelineDefn.Name()
			}
		} else {
			output.Status = constants.StateFinished
		}

		evalContext, err = execution.AddStepPrimitiveOutputAsResults(stepDefn.GetName(), output, evalContext)
		if err != nil {
			// catastrophic error - raise pipeline failed straight away
			slog.Error("Error adding step primitive output as results", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, err)
			return
		}

		// We have some special steps that need to be handled differently:
		// Pipeline Step -> launch a new pipeline
		// Input Step -> waiting for external event to resume the pipeline
		shouldReturn := specialStepHandler(ctx, stepDefn, cmd, evalContext, h)
		if shouldReturn {
			return
		}

		if output.HasErrors() {
			// check if we need to ignore the error
			errorConfig, diags := stepDefn.GetErrorConfig(evalContext, true)
			if diags.HasErrors() {
				slog.Error("Error getting error config", "error", diags)
				output.Status = constants.StateFailed
				output.FailureMode = constants.FailureModeFatal
			} else if errorConfig != nil && errorConfig.Ignore != nil && *errorConfig.Ignore {
				output.Status = constants.StateFailed
				output.FailureMode = constants.FailureModeIgnored
			} else {
				output.FailureMode = constants.FailureModeStandard
			}
		} else {
			output.Status = constants.StateFinished
		}

		if output.Status == constants.StateFinished && stepDefn.GetType() == schema.BlockTypeInput && (o.IsServerMode || primitive.IsInputRouted()) {
			slog.Info("input step started, waiting for external response", "step", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID, "executionID", cmd.Event.ExecutionID)
			raisePipelinePlannedFromStepStart(stepDefn, cmd, h.EventBus)
			return
		}

		if output.Status == constants.StateFinished && stepDefn.GetType() == schema.BlockTypePipelineStepMessage && primitive.IsInputRouted() {
			slog.Info("routed message step started, waiting for external confirmation/response", "step", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID, "executionID", cmd.Event.ExecutionID)
			raisePipelinePlannedFromStepStart(stepDefn, cmd, h.EventBus)
			return
		}

		// Only calculate the step output if there are no errors or if the error is ignored. Either way it will end up
		// with output.Status == constants.StateFinished
		if output.Status == constants.StateFinished || output.FailureMode == constants.FailureModeIgnored {
			// If there's a for_each in the step definition, we need to insert the "each" magic variable
			// so the output can refer to it
			evalContext, stepOutput, err = calculateStepConfiguredOutput(ctx, stepDefn, evalContext, cmd.StepForEach, stepOutput)

			// If there's an error calculating the output, we need to fail the step, the ignored error directive will be ignored
			// and the retry directive will be ignored as well
			if err != nil {
				if !perr.IsPerr(err) {
					err = perr.InternalWithMessage(err.Error())
				}

				// Append the error and set the state to failed
				output.Status = constants.StateFailed
				output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
				output.Errors = append(output.Errors, resources.StepError{
					PipelineExecutionID: cmd.PipelineExecutionID,
					StepExecutionID:     cmd.StepExecutionID,
					Pipeline:            pipelineDefn.Name(),
					Step:                cmd.StepName,
					Error:               err.(perr.ErrorModel),
				})

			}
		}

		// All other primitives finish immediately.
		endStep(ex, cmd, output, stepOutput, stepDefn, evalContext, ctx, h.EventBus)

	}(ctx, c, h)

	return nil
}

// This should only be called by input steps. It raises a pipeline planned event which in turn will do a regular check
// to see if the pipeline needs to be automatically paused
func raisePipelinePlannedFromStepStart(stepDefn resources.PipelineStep, cmd *event.StepStart, eventBus FpEventBus) {
	if stepDefn.GetType() != schema.BlockTypePipelineStepInput {
		return
	}

	go func() {
		e := event.PipelinePlanned{
			Event:     event.NewFlowEvent(cmd.Event),
			NextSteps: []resources.NextStep{},
		}

		e.PipelineExecutionID = cmd.PipelineExecutionID

		err := eventBus.Publish(context.Background(), &e)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
	}()

}

// This function mutates stepOutput
//
// https://github.com/turbot/flowpipe/issues/419
//
// Evaluation error, i.e. calculating the output, it fails the step and the retry and ignore error directives are not followed.
//
// The way this function is returned, whatever output currently calculated will be returned.
func calculateStepConfiguredOutput(ctx context.Context, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext, cmdStepForEach *resources.StepForEach, stepOutput map[string]interface{}) (*hcl.EvalContext, map[string]interface{}, error) {
	for _, outputConfig := range stepDefn.GetOutputConfig() {
		if outputConfig.UnresolvedValue != nil {

			stepForEach := stepDefn.GetForEach()
			if stepForEach != nil {
				evalContext = execution.AddEachForEach(cmdStepForEach, evalContext)
			}

			ctyValue, diags := outputConfig.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 && diags.HasErrors() {
				slog.Error("Error calculating output on step start", "error", diags)

				err := error_helpers.HclDiagsToError(stepDefn.GetName(), diags)
				return evalContext, stepOutput, err
			}

			goVal, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				slog.Error("Error converting cty to go", "error", err)

				return evalContext, stepOutput, err
			}

			stepOutput[outputConfig.Name] = goVal
		} else {
			stepOutput[outputConfig.Name] = outputConfig.Value
		}
	}

	return evalContext, stepOutput, nil
}

// If it's a pipeline step, we need to do something else, we we need to start
// a new pipeline execution for the child pipeline
// If it's an input step, we can't complete the step until the API receives the input's answer
func specialStepHandler(ctx context.Context, stepDefn resources.PipelineStep, cmd *event.StepStart, evalCtx *hcl.EvalContext, h StepStartHandler) bool {

	if stepDefn.GetType() == schema.AttributeTypePipeline {
		args := resources.Input{}
		if cmd.StepInput[schema.AttributeTypeArgs] != nil {
			args = cmd.StepInput[schema.AttributeTypeArgs].(map[string]interface{})
		}

		// Validate the param before we start the nested param
		nestedPipelineName, ok := cmd.StepInput[schema.AttributeTypePipeline].(string)
		if !ok {
			slog.Error("Unable to get pipeline name from the step input")
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, perr.InternalWithMessage("Unable to get pipeline name from the step input"))
			return true
		}

		currentStepPipeline := stepDefn.GetPipeline()
		if currentStepPipeline == nil {
			slog.Error("Unable to get pipeline from step definition")
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, perr.InternalWithMessage("Unable to get pipeline from step definition"))
			return true
		}

		currentStepMod := currentStepPipeline.GetMod()
		if currentStepMod == nil {
			slog.Error("Unable to get mod from step definition")
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, perr.InternalWithMessage("Unable to get mod from step definition"))
			return true
		}

		nestedPipelineModFullVersion := cmd.StepInput["mod_full_version"].(string)
		pipelineDefnToCall, err := db.GetPipelineWithModFullVersion(nestedPipelineModFullVersion, nestedPipelineName)

		if err != nil {
			slog.Error("Unable to get pipeline " + nestedPipelineName + " from cache", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, err)
			return true
		}

		errs := parse.ValidateParams(pipelineDefnToCall, args, evalCtx)

		if len(errs) > 0 {
			slog.Error("Failed validating pipeline param", "errors", errs)
			// just pick the first error
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, errs[0])
			return true
		}

		e, err := event.NewStepPipelineStarted(
			event.ForStepStart(cmd),
			event.WithNewChildPipelineExecutionID(),
			event.WithChildPipeline(cmd.StepInput[schema.AttributeTypePipeline].(string), args))
		e.ChildPipelineName = nestedPipelineName
		e.ChildPipelineModFullVersion = pipelineDefnToCall.GetMod().CacheKey()

		if cmd.StepForEach != nil {
			e.Key = cmd.StepForEach.Key
		} else {
			e.Key = "0"
		}

		if err != nil {
			raisePipelineFailedEventFromPipelineStepStart(ctx, h.EventBus, cmd, err)
			return true
		}

		err = h.EventBus.Publish(ctx, e)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return true
	}

	return false
}

func EndStepFromApi(ex *execution.ExecutionInMemory, stepExecution *execution.StepExecution, pipelineDefn *resources.Pipeline, stepDefn resources.PipelineStep, output *resources.Output, eventBus FpEventBus) error {

	stepStartCmdRecreated := event.StepStart{
		Event: &event.Event{
			ExecutionID: ex.ID,
			CreatedAt:   time.Now(),
		},
		PipelineExecutionID: stepExecution.PipelineExecutionID,
		StepExecutionID:     stepExecution.ID,
		StepName:            stepExecution.Name,
		StepInput:           stepExecution.Input,

		StepForEach: stepExecution.StepForEach,
		StepLoop:    stepExecution.StepLoop,
		StepRetry:   stepExecution.StepRetry,
	}

	pe := ex.PipelineExecutions[stepExecution.PipelineExecutionID]

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
	if err != nil {
		slog.Error("Error building eval context (end step handler)", "error", err)
		return err
	}

	evalContext, err = execution.AddStepPrimitiveOutputAsResults(stepDefn.GetName(), output, evalContext)
	if err != nil {
		slog.Error("Error adding step primitive output as results", "error", err)
		return err
	}

	endStep(ex, &stepStartCmdRecreated, output, nil, stepDefn, evalContext, context.Background(), eventBus)

	return nil
}

func endStep(ex *execution.ExecutionInMemory, cmd *event.StepStart, output *resources.Output, stepOutput map[string]interface{}, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext, ctx context.Context, eventBus FpEventBus) {

	// we need this to calculate the throw and loop, so might as well add it here for convenience
	endStepEvalContext, err := execution.AddStepCalculatedOutputAsResults(stepDefn.GetName(), stepOutput, &cmd.StepInput, evalContext)
	if err != nil {
		// catastrophic error - raise pipeline failed straight away
		slog.Error("Error adding step output as results", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, eventBus, cmd, err)
		return
	}

	// errors are handled in the following order: https://github.com/turbot/flowpipe/issues/419
	errorFromThrow := false
	stepError, err := calculateThrow(ctx, stepDefn, endStepEvalContext)
	if err != nil {
		// non-catasthropic error, fail the step, ignore the "retry" or "ignore" directive
		slog.Error("Error calculating throw", "error", err)

		if !perr.IsPerr(err) {
			err = perr.InternalWithMessage(err.Error())
		}

		// Append the error and set the state to failed
		output.Status = constants.StateFailed
		output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
		output.Errors = append(output.Errors, resources.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			Pipeline:            stepDefn.GetPipelineName(),
			StepExecutionID:     cmd.StepExecutionID,
			Step:                cmd.StepName,
			Error:               err.(perr.ErrorModel),
		})
	} else if stepError != nil {
		slog.Debug("Step error calculated from throw", "error", stepError)
		errorFromThrow = true
		output.Status = constants.StateFailed
		output.Errors = append(output.Errors, resources.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			Pipeline:            stepDefn.GetPipelineName(),
			StepExecutionID:     cmd.StepExecutionID,
			Step:                cmd.StepName,
			Error:               *stepError,
		})
	}

	if output.Status == constants.StateFailed && output.FailureMode != constants.FailureModeFatal {
		var stepRetry *resources.StepRetry
		var diags hcl.Diagnostics

		// Retry does not catch throw, so do not calculate the "retry" and automatically set the stepRetry to nil
		// to "complete" the error
		if !errorFromThrow {
			stepRetry, diags = calculateRetry(ctx, cmd.StepRetry, stepDefn, endStepEvalContext)
			if len(diags) > 0 {
				slog.Error("Error calculating retry", "diags", diags)

				err := error_helpers.HclDiagsToError(stepDefn.GetName(), diags)
				output.Status = constants.StateFailed
				output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
				output.Errors = append(output.Errors, resources.StepError{
					PipelineExecutionID: cmd.PipelineExecutionID,
					Pipeline:            stepDefn.GetPipelineName(),
					StepExecutionID:     cmd.StepExecutionID,
					Step:                cmd.StepName,
					Error:               err.(perr.ErrorModel),
				})
			}
		}

		if stepRetry != nil {
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry.Input = &cmd.StepInput
		} else {
			retryIndex := 0
			if cmd.StepRetry != nil {
				retryIndex = cmd.StepRetry.Count
			}
			// we have exhausted our retry, do not try to loop call step finish immediately
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry = &resources.StepRetry{
				Count:          retryIndex,
				RetryCompleted: true,
			}
		}

		// Now we have to check again if the error is ignored. Earlier in the process we checked if the error is ignored IF there's an error
		// however that may have changed now because the primitive may not have failed but there's a failure now due to the throw
		//
		//
		// check if we need to ignore the error
		errorConfig, diags := stepDefn.GetErrorConfig(evalContext, true)
		if diags.HasErrors() {
			slog.Error("Error getting error config", "error", diags)
			output.Status = constants.StateFailed
			output.FailureMode = constants.FailureModeFatal
		} else if errorConfig != nil && errorConfig.Ignore != nil && *errorConfig.Ignore {
			output.Status = constants.StateFailed
			output.FailureMode = constants.FailureModeIgnored
		} else {
			output.FailureMode = constants.FailureModeStandard
		}

		e, err := event.NewStepFinishedFromStepStart(cmd, output, stepOutput, cmd.StepLoop)
		if err != nil {
			slog.Error("Error creating Pipeline Step Finished event", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, eventBus, cmd, err)
			return
		}
		e.StepRetry = stepRetry

		// Don't forget to carry whatever the current loop config is
		e.StepLoop = cmd.StepLoop
		err = eventBus.Publish(ctx, e)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return

	}

	loopConfig := stepDefn.GetLoopConfig()

	var stepLoop *resources.StepLoop

	// Loop is calculated last, so it needs to respect the IF block evaluation
	if !helpers.IsNil(loopConfig) && cmd.NextStepAction != resources.NextStepActionSkip {
		var err error
		stepLoop, err = calculateLoop(ctx, ex, loopConfig, cmd.StepLoop, cmd.StepForEach, stepDefn, endStepEvalContext)
		if err != nil {
			slog.Error("Error calculating loop", "error", err)
			// Failure from loop calculation ignores ignore = true and retry block
			if !perr.IsPerr(err) {
				err = perr.InternalWithMessage(err.Error())
			}

			output.Status = constants.StateFailed
			output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
			output.Errors = append(output.Errors, resources.StepError{
				PipelineExecutionID: cmd.PipelineExecutionID,
				Pipeline:            stepDefn.GetPipelineName(),
				StepExecutionID:     cmd.StepExecutionID,
				Step:                cmd.StepName,
				Error:               err.(perr.ErrorModel),
			})
		}
	}

	e, err := event.NewStepFinishedFromStepStart(cmd, output, stepOutput, stepLoop)
	if err != nil {
		slog.Error("Error creating Pipeline Step Finished event", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, eventBus, cmd, err)
		return
	}

	err = eventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}
}

func raisePipelineFailedEventFromPipelineStepStart(ctx context.Context, eventBus FpEventBus, cmd *event.StepStart, originalError error) {
	err := eventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, originalError)))
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}
}

// This function returns 2 error. The first error is the result of the "throw" calculation, the second
// error is system error that should lead directly to pipeline fail
func calculateThrow(ctx context.Context, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext) (*perr.ErrorModel, error) {
	throwConfigs := stepDefn.GetThrowConfig()

	if len(throwConfigs) == 0 {
		return nil, nil
	}

	// We want the client to resolve the throw configuration. This to avoid failing on the subsequent throw if an
	// earlier throw is executed/
	//
	// For example: 3 throw configuration. If the first throw condition is met, then there's no reason we should evaluate the subsequent
	// throw conditions, let alone failing their evaluation.
	for _, throwConfig := range throwConfigs {
		resolvedThrowConfig, diags := throwConfig.Resolve(evalContext)
		if len(diags) > 0 {
			slog.Error("Error resolving throw config", "error", diags)
			return nil, error_helpers.HclDiagsToError(stepDefn.GetName(), diags)
		}

		if resolvedThrowConfig.If != nil && *resolvedThrowConfig.If {
			var message string
			if resolvedThrowConfig.Message != nil {
				message = *resolvedThrowConfig.Message
			} else {
				message = "User defined error"
			}
			stepErr := perr.UserDefinedWithMessage(message)
			return &stepErr, nil
		}
	}

	return nil, nil
}

func calculateRetry(ctx context.Context, stepRetry *resources.StepRetry, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext) (*resources.StepRetry, hcl.Diagnostics) {
	// we have error, check the if there's a retry block
	retryConfig, diags := stepDefn.GetRetryConfig(evalContext, true)

	if len(diags) > 0 {
		return nil, diags
	}

	if helpers.IsNil(retryConfig) {
		// there's no retry config ... nothing to retry
		return nil, hcl.Diagnostics{}
	}

	// if step retry == nil means this is the first time we encountered this issue
	if stepRetry == nil {
		stepRetry = &resources.StepRetry{
			Count: 0,
		}
	}

	stepRetry.Count = stepRetry.Count + 1

	maxAttempts, _, _, _ := retryConfig.ResolveSettings()

	// Max attempts include the first attempt (before the retry), so we need to reduce it by 1
	if stepRetry.Count > (maxAttempts - 1) {
		// we have exhausted all retries, we need to fail the pipeline
		return nil, hcl.Diagnostics{}
	}

	return stepRetry, hcl.Diagnostics{}
}

func calculateLoop(ctx context.Context, ex *execution.ExecutionInMemory, loopConfig resources.LoopDefn, stepLoop *resources.StepLoop, stepForEach *resources.StepForEach, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext) (*resources.StepLoop, error) {

	// If this is the first iteration of the loop, the cmd.StepLoop should be nil
	// thus the loop.index in the evaluation context should be 0
	//
	// this allows evaluation such as:
	/*

		step "echo" "echo" {
			text = "foo"

			loop {
				if = loop.index < 2
			}
		}
	**/
	//
	// Because this IF is still part of the Iteration 0 loop.index should be 0
	loopEvalContext := execution.AddLoop(stepLoop, evalContext)

	// We need to evaluate the "until" attribute separately than the rest of the loop body. Consider the following example:
	//
	// step "http" "http_list_pagination" {
	// 	url    = "https://some.url.com"
	//
	// 	loop {
	// 	  until = lookup(result.response_body, "next", null) == null
	// 	  url   = lookup(result.response_body, "next", "")
	// 	}
	// }
	//
	// The url may be invalid when until is reached, so we need to evaluate the until attribute first, independently,
	// then evaluate the rest of the loop block

	untilReached, diags := loopConfig.ResolveUntil(loopEvalContext)

	// We have to evaluate the loop body here before the index is incremented to determine if the loop should run
	// we will have to re-evaluate the loop body again after the index is incremented to get the correct values

	if len(diags) > 0 {
		return nil, perr.InternalWithMessage("error evaluating until attribute: " + diags.Error())
	}

	// We have to indicate here (before raising the step finish) that this is part of the loop that should be executing, i.e. the step is not actually
	// "finished" yet.
	//
	// Unlike the for_each where we know that there are n number of step executions and the planner launched them all at once, the loop is different.
	//
	// The planner has no idea that the step is not yet finished. We have to tell the planner here that it needs to launch another step execution

	// until has been reached so nothing to do
	if untilReached {
		newStepLoop := stepLoop
		// complete the loop
		// input is not required here
		stepLoop.Input = nil
		stepLoop.LoopCompleted = true
		return newStepLoop, nil
	}

	// Start with 1 because when we get here the first time, it was the 1st iteration of the loop (index = 0)
	//
	// Unlike the previous evaluation, we are not calculating the input for the NEXT iteration of the loop, so we need to increment the index,
	// do not change the currentIndex to 0
	currentIndex := 1
	if stepLoop != nil {
		previousIndex := stepLoop.Index
		currentIndex = previousIndex + 1
	}

	newStepLoop := &resources.StepLoop{
		Index: currentIndex,
		// Input: &newInput,
	}

	// first we need to evaluate the input for the step, this is to support:
	/*
		step "echo" "echo" {
			text = "iteration: ${loop.index}"
		}
	**/
	// for each of the loop iteration the index changes, so we have to re do it again
	// ensure that we also have the "each" variable here
	evalContext = execution.AddLoop(newStepLoop, evalContext)
	evalContext = execution.AddEachForEach(stepForEach, evalContext)

	reevaluatedInput, err := stepDefn.GetInputs(evalContext)
	if err != nil {
		slog.Error("Error re-evaluating inputs for step", "error", err)
		return nil, perr.InternalWithMessage("error re-evaluating inputs for step: " + err.Error())

	}

	// get the new input
	// ! we have to re add the "old" loop value, because the loopDefn should be evaluated using the old index
	// ! this is confusing .. so please read on:
	/**
	step "transform" "foo" {
		value = "loop: ${loop.index}"

		loop {
			until = loop.index < 2
			value = "new value: ${loop.index}"
		}
	}
	*/
	//
	// The loop in the step above is evaluated using the "prior" evalContext, however at this point of the execution
	// the evalContext's loop has been updated using the new index (+1) ... so we need to reverse the increment and
	// put it back to the old value.
	//
	// Check TestSimpleLoop to gain more understanding about this odd code.
	//
	evalContext = execution.AddLoop(stepLoop, evalContext)

	newInput, err := loopConfig.UpdateInput(reevaluatedInput, evalContext)
	if err != nil {
		return nil, perr.InternalWithMessage("error updating input for loop: " + err.Error())
	}

	newStepLoop.Input = &newInput
	return newStepLoop, nil
}
