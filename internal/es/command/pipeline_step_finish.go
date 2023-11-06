package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStepFinishHandler CommandHandler

func (h PipelineStepFinishHandler) HandlerName() string {
	return "command.pipeline_step_finish"
}

func (h PipelineStepFinishHandler) NewCommand() interface{} {
	return &event.PipelineStepFinish{}
}

// There's only one use case for this, which is to handle the "Pipeline Step" finish command.
//
// Pipeline Step = step that launches another pipeline.
//
// This command is NOT to to be confused with the handling of the "Pipeline Step" operation. That flow:
// Pipeline Step Start command -> Pipeline Step Finish *event*
func (h PipelineStepFinishHandler) Handle(ctx context.Context, c interface{}) error {

	logger := fplog.Logger(ctx)
	cmd, ok := c.(*event.PipelineStepFinish)
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

	e, err := event.NewPipelineStepFinished(event.ForPipelineStepFinish(cmd))
	e.StepOutput = stepOutput

	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
