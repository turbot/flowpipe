package command

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
)

type PipelineFailHandler CommandHandler

func (h PipelineFailHandler) HandlerName() string {
	return "command.pipeline_fail"
}

func (h PipelineFailHandler) NewCommand() interface{} {
	return &event.PipelineFail{}
}

func (h PipelineFailHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	cmd, ok := c.(*event.PipelineFail)
	if !ok {
		logger.Error("pipeline_fail handler expected PipelineFail event", "event", c)
		return fperr.BadRequestWithMessage("pipeline_fail handler expected PipelineFail event")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		logger.Error("pipeline_fail error constructing execution", "error", err)
		return err
	}

	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		logger.Error("Pipeline definition not found", "error", err)
		return err
	}

	// calculate as many output as ppossible
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
				logger.Info("Error calculating output during pipeline_fail during pipeline_fail event", "error", err)
				// do not fail, continue to the next output
				continue
			}
			val, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				logger.Error("Error converting cty value to Go value for output calculation during pipeline_fail event", "error", err)
				// do not fail, continue to the next output
				continue
			}
			outputBlock[output.Name] = val
		}
		output = outputBlock
	}

	return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFail(cmd, output)))
}
