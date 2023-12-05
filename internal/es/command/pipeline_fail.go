package command

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineFailHandler CommandHandler

func (h PipelineFailHandler) HandlerName() string {
	return execution.PipelineFailCommand.HandlerName()
}

func (h PipelineFailHandler) NewCommand() interface{} {
	return &event.PipelineFail{}
}

func (h PipelineFailHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineFail)
	if !ok {
		slog.Error("pipeline_fail handler expected PipelineFail event", "event", c)
		return perr.BadRequestWithMessage("pipeline_fail handler expected PipelineFail event")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		slog.Error("pipeline_fail error constructing execution", "error", err)
		return err
	}

	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		slog.Error("Pipeline definition not found", "error", err)
		return err
	}

	// 2023-12-05: do not calculate output if pipeline fails
	output := make(map[string]any, 1)

	// Collect all the step output, but don't also add the error in the cmd/event
	var pipelineErrors []modconfig.StepError
	if cmd.Error != nil {
		pipelineErrors = append(pipelineErrors, *cmd.Error)
	}

	for _, stepExecution := range pe.StepExecutions {
		stepDefn := pipelineDefn.GetStep(stepExecution.Name)
		if err != nil {
			slog.Error("Error getting step definition during pipeline_fail event", "error", err)
			// do not fail, continue to the next step, we are already in pipeline_fail event, what else can we do here?
			continue
		}
		if stepExecution.Output.HasErrors() && (stepDefn.GetErrorConfig() != nil && !stepDefn.GetErrorConfig().Ignore) {
			pipelineErrors = append(pipelineErrors, stepExecution.Output.Errors...)
		}
	}

	pipelineFailedEvent := event.NewPipelineFailedFromPipelineFail(cmd, output, pipelineErrors)
	return h.EventBus.Publish(ctx, pipelineFailedEvent)
}
