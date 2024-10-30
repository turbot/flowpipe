package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/resources"
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

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	executionID := cmd.Event.ExecutionID
	ex, _, err := execution.GetPipelineDefnFromExecution(executionID, cmd.PipelineExecutionID)
	if err != nil {
		// catasthropic failure here
		return err
	}
	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	// 2023-12-05: do not calculate output if pipeline fails
	output := make(map[string]any, 1)

	// Collect all the step output, but don't also add the error in the cmd/event
	var pipelineErrors []resources.StepError
	if cmd.Error != nil {
		pipelineErrors = append(pipelineErrors, *cmd.Error)
	}

	for _, stepExecution := range pe.StepExecutions {
		if stepExecution.Output.HasErrors() {
			if stepExecution.StepRetry != nil && !stepExecution.StepRetry.RetryCompleted {
				// Don't add to pipeline errors if it's not the final retry
				continue
			}

			if stepExecution.Output.FailureMode == constants.FailureModeIgnored {
				continue
			}

			pipelineErrors = append(pipelineErrors, stepExecution.Output.Errors...)
		}
	}

	pipelineFailedEvent := event.NewPipelineFailedFromPipelineFail(cmd, output, pipelineErrors)
	return h.EventBus.Publish(ctx, pipelineFailedEvent)
}
