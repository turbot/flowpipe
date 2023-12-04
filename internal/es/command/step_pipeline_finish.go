package command

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"log/slog"
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

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		slog.Error("Error loading pipeline execution", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		slog.Error("Error loading pipeline definition", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]
	stepExecution := pex.StepExecutions[cmd.StepExecutionID]
	stepDefn := pipelineDefn.GetStep(stepExecution.Name)

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
	if err != nil {
		slog.Error("Error building eval context", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	stepOutput := make(map[string]interface{})

	// Calculate the configured step output
	//
	// Ignore the merging here, the nested pipeline output is also called "output", but that merging is done later
	// when we build the evalContext.
	//
	// As long as they are in 2 different property: Output (native output, happens also to be called "output" for pipeline step) and StepOutput (also referred to configured step output)
	// we will be OK
	for _, outputConfig := range stepDefn.GetOutputConfig() {
		if outputConfig.UnresolvedValue != nil {

			stepForEach := stepDefn.GetForEach()
			if stepForEach != nil {
				evalContext = execution.AddEachForEach(cmd.StepForEach, evalContext)
			}

			ctyValue, diags := outputConfig.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 && diags.HasErrors() {
				slog.Error("Error calculating output on step start", "error", diags)
				stepOutput[outputConfig.Name] = "Error calculating output " + diags.Error()
				continue
			}

			goVal, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				slog.Error("Error converting cty value to Go value for output calculation", "error", err)
				stepOutput[outputConfig.Name] = "Error calculating output " + err.Error()
				continue
			}
			stepOutput[outputConfig.Name] = goVal
		} else {
			stepOutput[outputConfig.Name] = outputConfig.Value
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
	endStepEvalContext, err := execution.AddStepOutputAsResults(stepDefn.GetName(), cmd.Output, stepOutput, evalContext)

	if err != nil {
		slog.Error("Error adding step output as results", "error", err)
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	stepError, err := calculateThrow(ctx, stepDefn, endStepEvalContext)
	errorFromThrow := false
	if err != nil {
		slog.Error("Error calculating throw", "error", err)
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	if stepError != nil {
		slog.Debug("Step error calculated from throw", "error", stepError)
		errorFromThrow = true
		cmd.Output.Status = "failed"
		cmd.Output.Errors = append(cmd.Output.Errors, modconfig.StepError{
			PipelineExecutionID: cmd.PipelineExecutionID,
			StepExecutionID:     cmd.StepExecutionID,
			Step:                stepDefn.GetName(),
			Error:               *stepError,
		})
	}

	if cmd.Output.Status == "failed" {
		var stepRetry *modconfig.StepRetry
		var diags hcl.Diagnostics
		// Retry does not catch throw, so do not calculate the "retry" and automatically set the stepRetry to nil
		// to "complete" the error
		if !errorFromThrow {
			stepRetry, diags = calculateRetry(ctx, cmd.StepRetry, stepDefn, endStepEvalContext)

			if len(diags) > 0 {
				slog.Error("Error calculating retry", "diags", diags)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, error_helpers.HclDiagsToError(stepDefn.GetName(), diags)))
				if err2 != nil {
					slog.Error("Error publishing event", "error", err2)
				}
				return nil
			}
		}

		if stepRetry != nil {
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry.Input = &cmd.StepInput
		} else {
			// means we need to retry, ignore the loop right now, we need to retry first to clear the error
			stepRetry = &modconfig.StepRetry{
				RetryCompleted: true,
			}
		}

		e, err := event.NewStepFinished(event.ForPipelineStepFinish(cmd))
		if err != nil {
			slog.Error("Error creating Pipeline Step Finished event", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
		}
		e.StepRetry = stepRetry
		// e.StepLoop = cmd.StepLoop
		err = h.EventBus.Publish(ctx, e)

		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
		}
		return nil
	}

	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	var stepLoop *modconfig.StepLoop
	if loopBlock != nil {
		var err error
		stepLoop, err = calculateLoop(ctx, ex, loopBlock, cmd.StepLoop, cmd.StepForEach, stepDefn, endStepEvalContext)
		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
			if err2 != nil {
				slog.Error("Error publishing event", "error", err2)
			}
		}
	}

	e, err := event.NewStepFinished(event.ForPipelineStepFinish(cmd))
	e.StepLoop = stepLoop

	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromStepPipelineFinish(cmd, err))
	}
	e.StepOutput = stepOutput

	return h.EventBus.Publish(ctx, e)
}
