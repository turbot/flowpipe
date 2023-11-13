package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
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

	logger := fplog.Logger(ctx)
	cmd, ok := c.(*event.StepPipelineFinish)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineStepFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineStepFinish")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		logger.Error("Error loading pipeline execution", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
		if err2 != nil {
			logger.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		logger.Error("Error loading pipeline definition", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
		if err2 != nil {
			logger.Error("Error publishing event", "error", err2)
		}
		return nil
	}

	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]
	stepExecution := pex.StepExecutions[cmd.StepExecutionID]
	stepDefn := pipelineDefn.GetStep(stepExecution.Name)

	evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
	if err != nil {
		logger.Error("Error building eval context", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
		if err2 != nil {
			logger.Error("Error publishing event", "error", err2)
		}
		return nil
	}
	stepOutput := make(map[string]interface{})

	if cmd.Output.Status == "failed" {
		stepRetry := calculateRetry(ctx, cmd.StepRetry, stepDefn)

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
			logger.Error("Error creating Pipeline Step Finished event", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
		}
		e.StepRetry = stepRetry
		// e.StepLoop = cmd.StepLoop
		err = h.EventBus.Publish(ctx, e)

		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
		}
		return nil
	}

	// Calculate the configured step output
	//
	// Ignore the merging here, the nested pipeline output is also called "output", but that merging is done later
	// when we build the evalContext.
	//
	// As long as they are in 2 different property: Output (native output, happens also to be called "output" for pipeline step) and StepOutput (also referred to configured step output)
	// we will be OK
	if !cmd.Output.HasErrors() {
		for _, outputConfig := range stepDefn.GetOutputConfig() {
			if outputConfig.UnresolvedValue != nil {

				stepForEach := stepDefn.GetForEach()
				if stepForEach != nil {
					evalContext = execution.AddEachForEach(cmd.StepForEach, evalContext)
				}

				ctyValue, diags := outputConfig.UnresolvedValue.Value(evalContext)
				if len(diags) > 0 && diags.HasErrors() {
					logger.Error("Error calculating output on step start", "error", diags)
					err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
					if err2 != nil {
						logger.Error("Error publishing event", "error", err2)
					}
					return nil
				}

				goVal, err := hclhelpers.CtyToGo(ctyValue)
				if err != nil {
					logger.Error("Error converting cty value to Go value for output calculation", "error", err)
					err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
					if err2 != nil {
						logger.Error("Error publishing event", "error", err2)
					}
					return nil
				}
				stepOutput[outputConfig.Name] = goVal
			} else {
				stepOutput[outputConfig.Name] = outputConfig.Value
			}
		}
	}

	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	var stepLoop *modconfig.StepLoop
	if loopBlock != nil {
		var err error
		stepLoop, err = calculateLoop(ctx, loopBlock, cmd.StepLoop, cmd.StepForEach, stepDefn, evalContext, stepDefn.GetName(), cmd.Output, stepOutput)
		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
		}
	}

	e, err := event.NewStepFinished(event.ForPipelineStepFinish(cmd))
	e.StepLoop = stepLoop

	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
	}
	e.StepOutput = stepOutput

	return h.EventBus.Publish(ctx, e)
}
