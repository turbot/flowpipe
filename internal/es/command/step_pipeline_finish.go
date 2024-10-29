package command

import (
	"context"
	"github.com/turbot/flowpipe/internal/resources"
	"time"

	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
)

type StepPipelineFinishHandler CommandHandler

func (h StepPipelineFinishHandler) HandlerName() string {
	return execution.StepPipelineFinishCommand.HandlerName()
}

func (h StepPipelineFinishHandler) NewCommand() interface{} {
	return &event.StepPipelineFinish{}
}

// There's only one use case for this, which is to handle the "Pipeline Step" finish command.
//
// Pipeline Step = step that launches another pipeline.
//
// This command is NOT to to be confused with the handling of the "Pipeline Step" operation. That flow:
// Pipeline Step Start command -> Pipeline Step Finish *event*
func (h StepPipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.StepPipelineFinish)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineStepFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineStepFinish")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(cmd.Event.ExecutionID, cmd.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_plan: Error loading pipeline execution", "error", err)
		raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
		return nil
	}
	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	stepExecution := pex.StepExecutions[cmd.StepExecutionID]
	stepDefn := pipelineDefn.GetStep(stepExecution.Name)

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
	if err != nil {
		slog.Error("Error building eval context (pipeline finish handler)", "error", err)

		raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
		return nil
	}

	stepOutput := make(map[string]interface{})

	cmd.Output.Flowpipe = primitive.FlowpipeMetadataOutput(pex.StartTime, time.Now().UTC())

	if cmd.Output.Status == constants.StateFailed {
		errorConfig, diags := stepDefn.GetErrorConfig(evalContext, true)
		if diags.HasErrors() {
			slog.Error("Error getting error config", "error", diags)
			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeFatal
		} else if errorConfig != nil && errorConfig.Ignore != nil && *errorConfig.Ignore {
			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeIgnored
		} else {
			cmd.Output.FailureMode = constants.FailureModeStandard
		}
	} else {
		cmd.Output.Status = constants.StateFinished
	}

	// Calculate the configured step output
	//
	// Ignore the merging here, the nested pipeline output is also called "output", but that merging is done later
	// when we build the evalContext.
	//
	// As long as they are in 2 different property: Output (native output, happens also to be called "output" for pipeline step) and StepOutput (also referred to configured step output)
	// we will be OK
	if cmd.Output.Status == constants.StateFinished || cmd.Output.FailureMode == constants.FailureModeIgnored {
		evalContext, stepOutput, err = calculateStepConfiguredOutput(ctx, stepDefn, evalContext, cmd.StepForEach, stepOutput)
		// If there's an error calculating the output, we need to fail the step, the ignored error directive will be ignored
		// and the retry directive will be ignored as well
		if err != nil {
			if !perr.IsPerr(err) {
				err = perr.InternalWithMessage(err.Error())
			}

			// Append the error and set the state to failed
			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
			cmd.Output.Errors = append(cmd.Output.Errors, resources.StepError{
				PipelineExecutionID: cmd.PipelineExecutionID,
				StepExecutionID:     cmd.StepExecutionID,
				Pipeline:            pipelineDefn.Name(),
				Step:                stepDefn.GetName(),
				Error:               err.(perr.ErrorModel),
			})
		}
	}

	// We need this to calculate the throw and loop, so might as well add it here for convenience
	//
	// If there's an error calculating the eval context, we have 2 options:
	// 1) raise pipeline_failed event, or
	// 2) set the output as "failed" and raise step_finish event
	//
	// I can see there are merit for both. #2 is usually the right way because we can ignore error, however this type
	// of problem, e.g. building eval context failure due to clash in the step output, is a configuration error, so I think
	// it should raise pipeline_failed event directly
	endStepEvalContext, err := execution.AddStepPrimitiveOutputAsResults(stepDefn.GetName(), cmd.Output, evalContext)
	if err != nil {
		slog.Error("Error adding step primitive output as results", "error", err)
		raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
		return nil
	}

	endStepEvalContext, err = execution.AddStepCalculatedOutputAsResults(stepDefn.GetName(), stepOutput, &cmd.StepInput, endStepEvalContext)
	if err != nil {
		slog.Error("Error adding step calculated output as results", "error", err)
		raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
		return nil
	}

	stepError, err := calculateThrow(ctx, stepDefn, endStepEvalContext)
	errorFromThrow := false
	if err != nil {
		slog.Error("Error calculating throw", "error", err)
		// non-catasthropic error, fail the step, ignore the "retry" or "ignore" directive

		if !perr.IsPerr(err) {
			err = perr.InternalWithMessage(err.Error())
		}
		// Append the error and set the state to failed
		cmd.Output.Status = constants.StateFailed
		cmd.Output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
		cmd.Output.Errors = append(cmd.Output.Errors, resources.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			Pipeline:            stepDefn.GetPipelineName(),
			StepExecutionID:     cmd.StepExecutionID,
			Step:                stepDefn.GetName(),
			Error:               err.(perr.ErrorModel),
		})
	} else if stepError != nil {
		slog.Debug("Step error calculated from throw", "error", stepError)

		errorFromThrow = true
		cmd.Output.Status = constants.StateFailed
		cmd.Output.Errors = append(cmd.Output.Errors, resources.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			StepExecutionID:     cmd.StepExecutionID,
			Step:                stepDefn.GetName(),
			Error:               *stepError,
		})
	}

	if cmd.Output.Status == constants.StateFailed && cmd.Output.FailureMode != constants.FailureModeFatal {
		var stepRetry *resources.StepRetry
		var diags hcl.Diagnostics
		// Retry does not catch throw, so do not calculate the "retry" and automatically set the stepRetry to nil
		// to "complete" the error
		if !errorFromThrow {
			stepRetry, diags = calculateRetry(ctx, cmd.StepRetry, stepDefn, endStepEvalContext)

			if len(diags) > 0 {
				slog.Error("Error calculating retry", "diags", diags)

				err := error_helpers.HclDiagsToError(stepDefn.GetName(), diags)
				cmd.Output.Status = constants.StateFailed
				cmd.Output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
				cmd.Output.Errors = append(cmd.Output.Errors, resources.StepError{
					PipelineExecutionID: cmd.PipelineExecutionID,
					Pipeline:            stepDefn.GetPipelineName(),
					StepExecutionID:     cmd.StepExecutionID,
					Step:                stepDefn.GetName(),
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

			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry = &resources.StepRetry{
				Count:          retryIndex,
				RetryCompleted: true,
			}
		}

		errorConfig, diags := stepDefn.GetErrorConfig(evalContext, true)
		if diags.HasErrors() {
			slog.Error("Error getting error config", "error", diags)
			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeFatal
		} else if errorConfig != nil && errorConfig.Ignore != nil && *errorConfig.Ignore {
			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeIgnored
		} else {
			cmd.Output.FailureMode = constants.FailureModeStandard
		}

		e, err := event.NewStepFinished(event.ForPipelineStepFinish(cmd))
		if err != nil {
			slog.Error("Error creating Pipeline Step Finished event", "error", err)
			raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
			return nil
		}
		e.StepRetry = stepRetry
		// e.StepLoop = cmd.StepLoop
		err = h.EventBus.Publish(ctx, e)

		if err != nil {
			raisePipelineFailedFromStepPipelineFinishError(ctx, h, cmd, err)
			return nil
		}

		return nil
	}

	loopConfig := stepDefn.GetLoopConfig()
	var stepLoop *resources.StepLoop
	if loopConfig != nil {
		var err error
		stepLoop, err = calculateLoop(ctx, ex, loopConfig, cmd.StepLoop, cmd.StepForEach, stepDefn, endStepEvalContext)
		if err != nil {
			if !perr.IsPerr(err) {
				err = perr.InternalWithMessage(err.Error())
			}

			cmd.Output.Status = constants.StateFailed
			cmd.Output.FailureMode = constants.FailureModeFatal // this is a indicator that this step should be retried or error ignored
			cmd.Output.Errors = append(cmd.Output.Errors, resources.StepError{
				PipelineExecutionID: cmd.PipelineExecutionID,
				Pipeline:            stepDefn.GetPipelineName(),
				StepExecutionID:     cmd.StepExecutionID,
				Step:                stepDefn.GetName(),
				Error:               err.(perr.ErrorModel),
			})
		}
	}

	err = execution.ReleasePipelineExecutionStepSemaphore(cmd.PipelineExecutionID, stepDefn)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
	}

	e, err := event.NewStepFinished(event.ForPipelineStepFinish(cmd))
	e.StepLoop = stepLoop

	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
	}
	e.StepOutput = stepOutput

	return h.EventBus.Publish(ctx, e)
}

func raisePipelineFailedFromStepPipelineFinishError(ctx context.Context, h StepPipelineFinishHandler, cmd *event.StepPipelineFinish, err error) {
	publishError := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
	if publishError != nil {
		slog.Error("Error publishing event", "error", publishError)
	}
}
