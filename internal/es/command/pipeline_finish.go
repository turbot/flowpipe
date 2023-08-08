package command

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return "command.pipeline_finish"
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	cmd, ok := c.(*event.PipelineFinish)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineFinish", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineFinish")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	var output map[string]interface{}
	if len(pipelineDefn.Outputs) > 0 {
		outputBlock := map[string]interface{}{}

		// If all dependencies met, we then calculate the value of this output
		evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
		if err != nil {
			logger.Error("Error building eval context while calculating output", "error", err)
			return err
		}

		for _, output := range pipelineDefn.Outputs {
			// check if its dependencies have been met
			dependenciesMet := true
			for _, dep := range output.DependsOn {
				if !pe.IsStepComplete(dep) {
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
				err := pipeparser.DiagsToError("output", diags)
				logger.Error("Error calculating output", "error", err)
				return err
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

	return h.EventBus.Publish(ctx, &e)
}
