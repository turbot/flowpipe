package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineFinishHandler CommandHandler

var pipelineFinish = event.PipelineFinish{}

func (h PipelineFinishHandler) HandlerName() string {
	return pipelineFinish.HandlerName()
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	cmd, ok := c.(*event.PipelineFinish)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineFinish")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	var output map[string]interface{}
	if len(pipelineDefn.OutputConfig) > 0 {
		outputBlock := map[string]interface{}{}

		// If all dependencies met, we then calculate the value of this output
		evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
		if err != nil {
			logger.Error("Error building eval context while calculating output in pipeline_finish", "error", err)
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

		for _, output := range pipelineDefn.OutputConfig {
			// check if its dependencies have been met
			dependenciesMet := true
			for _, dep := range output.DependsOn {
				if !pex.IsStepComplete(dep) {
					dependenciesMet = false
					break
				}
			}
			// Dependencies not met, skip this output
			if !dependenciesMet {
				continue
			}
			ctyValue, diags := output.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 {
				err := error_helpers.HclDiagsToError("output", diags)
				logger.Error("Error calculating output on pipeline finish", "error", err)
				outputBlock[output.Name] = "Unable to calculate output " + output.Name + ": " + err.Error()
				continue
			}
			val, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				logger.Error("Error converting cty value to Go value", "error", err)
				return err
			}
			outputBlock[output.Name] = val
		}
		output = outputBlock
	}

	e, err := event.NewPipelineFinished(event.ForPipelineFinish(cmd, output))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	return h.EventBus.Publish(ctx, e)
}
