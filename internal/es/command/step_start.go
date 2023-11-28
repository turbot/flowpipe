package command

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
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

		logger := fplog.Logger(ctx)

		cmd, ok := c.(*event.StepStart)
		if !ok {
			logger.Error("invalid command type", "expected", "*event.StepStart", "actual", c)
			return
		}

		ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
		if err != nil {
			logger.Error("Error loading pipeline execution", "error", err)

			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
		if err != nil {
			logger.Error("Error loading pipeline definition", "error", err)

			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		stepDefn := pipelineDefn.GetStep(cmd.StepName)
		stepOutput := make(map[string]interface{})

		pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

		evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
		if err != nil {
			logger.Error("Error building eval context while calculating output", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		// Check if the step should be skipped. This is determined by the evaluation of the IF clause during the
		// pipeline_plan phase
		if cmd.NextStepAction == modconfig.NextStepActionSkip {
			output := &modconfig.Output{
				Status: "skipped",
			}

			endStep(ex, cmd, output, stepOutput, logger, h, stepDefn, evalContext, ctx)
			return
		}

		var output *modconfig.Output

		var primitiveError error
		switch stepDefn.GetType() {
		case schema.BlockTypePipelineStepExec:
			p := primitive.Exec{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
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
		case schema.BlockTypePipelineStepEcho:
			p := primitive.Echo{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepTransform:
			p := primitive.Transform{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepFunction:
			p := primitive.Function{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepContainer:
			p := primitive.Container{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepInput:
			p := primitive.Input{
				ExecutionID:         cmd.Event.ExecutionID,
				PipelineExecutionID: cmd.PipelineExecutionID,
				StepExecutionID:     cmd.StepExecutionID,
			}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		default:
			logger.Error("Unknown step type", "type", stepDefn.GetType())
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		if primitiveError != nil {
			logger.Error("primitive failed", "error", primitiveError)
			if output == nil {
				output = &modconfig.Output{}
			}
			if output.Errors == nil {
				output.Errors = []modconfig.StepError{}
			}

			output.Errors = append(output.Errors, modconfig.StepError{
				Error: perr.InternalWithMessage(primitiveError.Error()),
			})

		}

		// Decorate the errors
		if output.HasErrors() {
			output.Status = "failed"
			for i := 0; i < len(output.Errors); i++ {
				(output.Errors)[i].Step = cmd.StepName
				(output.Errors)[i].PipelineExecutionID = cmd.PipelineExecutionID
				(output.Errors)[i].StepExecutionID = cmd.StepExecutionID
				(output.Errors)[i].Pipeline = pipelineDefn.Name()
			}
		} else {
			output.Status = "finished"
		}

		// We have some special steps that need to be handled differently:
		// Pipeline Step -> launch a new pipeline
		// Input Step -> waiting for external event to resume the pipeline
		shouldReturn := specialStepHandler(ctx, stepDefn, cmd, h, logger)
		if shouldReturn {
			return
		}

		// calculate the output blocks
		// If there's a for_each in the step definition, we need to insert the "each" magic variable
		// so the output can refer to it
		evalContext, stepOutput, shouldReturn = calculateStepConfiguredOutput(ctx, stepDefn, evalContext, cmd, logger, h, err, stepOutput)
		if shouldReturn {
			return
		}

		// All other primitives finish immediately.
		endStep(ex, cmd, output, stepOutput, logger, h, stepDefn, evalContext, ctx)

	}(ctx, c, h)

	return nil
}

// This function mutates stepOutput
func calculateStepConfiguredOutput(ctx context.Context, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext, cmd *event.StepStart, logger *fplog.FlowpipeLogger, h StepStartHandler, err error, stepOutput map[string]interface{}) (*hcl.EvalContext, map[string]interface{}, bool) {
	for _, outputConfig := range stepDefn.GetOutputConfig() {
		if outputConfig.UnresolvedValue != nil {

			stepForEach := stepDefn.GetForEach()
			if stepForEach != nil {

				evalContext = execution.AddEachForEach(cmd.StepForEach, evalContext)
			}

			ctyValue, diags := outputConfig.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 && diags.HasErrors() {
				logger.Error("Error calculating output on step start", "error", diags)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return nil, stepOutput, true
			}

			goVal, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				logger.Error("Error converting cty value to Go value for output calculation", "error", err)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return nil, stepOutput, true
			}
			stepOutput[outputConfig.Name] = goVal
		} else {
			stepOutput[outputConfig.Name] = outputConfig.Value
		}
	}
	return evalContext, stepOutput, false
}

// If it's a pipeline step, we need to do something else, we we need to start
// a new pipeline execution for the child pipeline
// If it's an input step, we can't complete the step until the API receives the input's answer
func specialStepHandler(ctx context.Context, stepDefn modconfig.PipelineStep, cmd *event.StepStart, h StepStartHandler, logger *fplog.FlowpipeLogger) bool {

	if stepDefn.GetType() == schema.AttributeTypePipeline {
		args := modconfig.Input{}
		if cmd.StepInput[schema.AttributeTypeArgs] != nil {
			args = cmd.StepInput[schema.AttributeTypeArgs].(map[string]interface{})
		}

		e, err := event.NewStepPipelineStarted(
			event.ForStepStart(cmd),
			event.WithNewChildPipelineExecutionID(),
			event.WithChildPipeline(cmd.StepInput[schema.AttributeTypePipeline].(string), args))

		if cmd.StepForEach != nil {
			e.Key = cmd.StepForEach.Key
		} else {
			e.Key = "0"
		}

		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return true
		}

		err = h.EventBus.Publish(ctx, e)
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}

		return true
	} else if stepDefn.GetType() == schema.BlockTypeInput {

		logger.Info("input step started, waiting for external response", "step", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID, "executionID", cmd.Event.ExecutionID)
		return true
	}

	return false
}

func endStep(ex *execution.Execution, cmd *event.StepStart, output *modconfig.Output, stepOutput map[string]interface{}, logger *fplog.FlowpipeLogger, h StepStartHandler, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext, ctx context.Context) {

	// we need this to calculate the throw and loop, so might as well add it here for convenience
	endStepEvalContext, err := execution.AddStepOutputAsResults(stepDefn.GetName(), output, stepOutput, evalContext)
	if err != nil {
		logger.Error("Error adding step output as results", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
		return
	}

	// errors are handled in the following order:
	//
	// throw, in the order that they appear
	// retry
	// error
	errorFromThrow := false
	stepError, err := calculateThrow(ctx, stepDefn, endStepEvalContext)
	if err != nil {
		logger.Error("Error calculating throw", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
		return
	}

	if stepError != nil {
		logger.Debug("Step error calculated from throw", "error", stepError)
		errorFromThrow = true
		output.Status = "failed"
		output.Errors = append(output.Errors, modconfig.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			StepExecutionID:     cmd.StepExecutionID,
			Step:                cmd.StepName,
			Error:               *stepError,
		})
	}

	if output.Status == "failed" {
		var stepRetry *modconfig.StepRetry
		var diags hcl.Diagnostics

		// Retry does not catch throw, so do not calculate the "retry" and automatically set the stepRetry to nil
		// to "complete" the error
		if !errorFromThrow {
			stepRetry, diags = calculateRetry(ctx, cmd.StepRetry, stepDefn, endStepEvalContext)
			if len(diags) > 0 {
				logger.Error("Error calculating retry", "diags", diags)
				raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, error_helpers.HclDiagsToError(stepDefn.GetName(), diags), logger)
				return
			}
		}

		if stepRetry != nil {
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry.Input = &cmd.StepInput
		} else {
			// we have exhausted our retry, do not try to loop call step finish immediately
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry = &modconfig.StepRetry{
				RetryCompleted: true,
			}
		}

		e, err := event.NewStepFinishedFromStepStart(cmd, output, stepOutput, cmd.StepLoop)
		if err != nil {
			logger.Error("Error creating Pipeline Step Finished event", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
			return
		}
		e.StepRetry = stepRetry

		// Don't forget to carry whatever the current loop config is
		e.StepLoop = cmd.StepLoop
		err = h.EventBus.Publish(ctx, e)
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}
		return

	}

	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	var stepLoop *modconfig.StepLoop
	if loopBlock != nil {
		var err error
		stepLoop, err = calculateLoop(ctx, ex, loopBlock, cmd.StepLoop, cmd.StepForEach, stepDefn, endStepEvalContext)
		if err != nil {
			logger.Error("Error calculating loop", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
			return
		}
	}

	e, err := event.NewStepFinishedFromStepStart(cmd, output, stepOutput, stepLoop)
	if err != nil {
		logger.Error("Error creating Pipeline Step Finished event", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
		return
	}

	err = h.EventBus.Publish(ctx, e)
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}

func raisePipelineFailedEventFromPipelineStepStart(ctx context.Context, h StepStartHandler, cmd *event.StepStart, originalError error, logger *fplog.FlowpipeLogger) {
	err := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepStartToPipelineFailed(cmd, originalError)))
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}

// This function returns 2 error. The first error is the result of the "throw" calculation, the second
// error is system error that should lead directly to pipeline fail
func calculateThrow(ctx context.Context, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext) (*perr.ErrorModel, error) {
	logger := fplog.Logger(ctx)

	throwConfigs := stepDefn.GetThrowConfig()
	if len(throwConfigs) == 0 {
		return nil, nil
	}

	for _, throwConfig := range throwConfigs {
		throwDefn := modconfig.ThrowConfig{}

		if throwConfig.Unresolved {
			diags := gohcl.DecodeBody(throwConfig.UnresolvedBody, evalContext, &throwDefn)

			if len(diags) > 0 && diags.HasErrors() {
				logger.Error("Error calculating throw", "error", diags)
				return nil, perr.InternalWithMessage("error calculating throw: " + diags.Error())
			}
		} else {
			throwDefn = throwConfig
		}

		if throwDefn.If {
			var message string
			if throwDefn.Message != nil {
				message = *throwDefn.Message
			} else {
				message = "Unkonwn error"
			}
			stepErr := perr.BadRequestWithMessage(message)
			return &stepErr, nil
		}
	}

	return nil, nil
}

func calculateRetry(ctx context.Context, stepRetry *modconfig.StepRetry, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext) (*modconfig.StepRetry, hcl.Diagnostics) {
	// we have error, check the if there's a retry block
	retryConfig, diags := stepDefn.GetRetryConfig(evalContext)

	if len(diags) > 0 {
		return nil, diags
	}

	if helpers.IsNil(retryConfig) {
		// there's no retry config ... nothing to retry
		return nil, hcl.Diagnostics{}
	}

	// if step retry == nil means this is the first time we encountered this issue
	if stepRetry == nil {
		stepRetry = &modconfig.StepRetry{
			Count: 0,
		}
	}

	stepRetry.Count = stepRetry.Count + 1

	// Max attempts include the first attempt (before the retry), so we need to reduce it by 1
	if stepRetry.Count > (retryConfig.MaxAttempts - 1) {
		// we have exhausted all retries, we need to fail the pipeline
		return nil, hcl.Diagnostics{}
	}

	return stepRetry, hcl.Diagnostics{}
}

func calculateLoop(ctx context.Context, ex *execution.Execution, loopBlock hcl.Body, stepLoop *modconfig.StepLoop, stepForEach *modconfig.StepForEach, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext) (*modconfig.StepLoop, error) {

	logger := fplog.Logger(ctx)

	loopDefn := modconfig.GetLoopDefn(stepDefn.GetType())
	if loopDefn == nil {
		// We should never get here, because the loop block should have been validated
		logger.Error("Unknown loop type", "type", stepDefn.GetType())
		return nil, perr.InternalWithMessage("unkonwn loop type: " + stepDefn.GetType())
	}

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

	// We have to evaluate the loop body here before the index is incremented to determine if the loop should run
	// we will have to re-evaluate the loop body again after the index is incremented to get the correct values
	diags := gohcl.DecodeBody(loopBlock, loopEvalContext, loopDefn)
	if len(diags) > 0 {
		return nil, perr.InternalWithMessage("error decoding loop block: " + diags.Error())
	}

	// We have to indicate here (before raising the step finish) that this is part of the loop that should be executing, i.e. the step is not actually
	// "finished" yet.
	//
	// Unlike the for_each where we know that there are n number of step executions and the planner launched them all at once, the loop is different.
	//
	// The planner has no idea that the step is not yet finished. We have to tell the planner here that it needs to launch another step execution
	if !loopDefn.UntilReached() {
		// Start with 1 because when we get here the first time, it was the 1st iteration of the loop (index = 0)
		//
		// Unlike the previous evaluation, we are not calculating the input for the NEXT iteration of the loop, so we need to increment the index,
		// do not change the currentIndex to 0
		currentIndex := 1
		if stepLoop != nil {
			previousIndex := stepLoop.Index
			currentIndex = previousIndex + 1
		}

		newStepLoop := &modconfig.StepLoop{
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

		var err error
		evalContext, err = ex.AddCredentialsToEvalContext(evalContext, stepDefn)
		if err != nil {
			logger.Error("Error adding credentials to eval context", "error", err)
			return nil, err
		}

		reevaluatedInput, err := stepDefn.GetInputs(evalContext)
		if err != nil {
			logger.Error("Error re-evaluating inputs for step", "error", err)
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

		newInput, err := loopDefn.UpdateInput(reevaluatedInput, evalContext)
		if err != nil {
			return nil, perr.InternalWithMessage("error updating input for loop: " + err.Error())
		}

		newStepLoop.Input = &newInput
		return newStepLoop, nil
	}

	newStepLoop := stepLoop
	// complete the loop
	// input is not required here
	stepLoop.Input = nil
	stepLoop.LoopCompleted = true
	return newStepLoop, nil
}
