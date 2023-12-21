package handler

import (
	"context"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
)

type PipelineFailed EventHandler

func (h PipelineFailed) HandlerName() string {
	return execution.PipelineFailedEvent.HandlerName()
}

func (PipelineFailed) NewEvent() interface{} {
	return &event.PipelineFailed{}
}

func (h PipelineFailed) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineFailed)

	slog.Debug("pipeline_failed handler", "event", e)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		slog.Error("pipeline_failed error constructing execution", "error", err)
		return err
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		// We're already in a pipeline failed event handler
		slog.Error("pipeline_failed error getting parent step execution", "error", err)
		return err
	}

	if parentStepExecution != nil {
		cmd, err := event.NewStepPipelineFinish(
			event.ForPipelineFailed(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID),

			// If StepForEach is not nil, it indicates that this pipeline execution is part of
			// for_each steps
			event.WithStepForEach(parentStepExecution.StepForEach))

		cmd.StepRetry = parentStepExecution.StepRetry
		cmd.StepInput = parentStepExecution.Input
		cmd.StepLoop = parentStepExecution.StepLoop

		if err != nil {
			slog.Error("pipeline_failed error creating pipeline step finish event", "error", err)
			return err
		}

		return h.CommandBus.Send(ctx, cmd)
	}

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		slog.Error("Pipeline definition not found", "error", err)
		return err
	}

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutput(e.Event.CreatedAt, "pipeline", "failed"),
			e.Event.ExecutionID, pipelineDefn.PipelineName)
		p.Errors = e.Errors
		p.Output = e.PipelineOutput
		output.RenderServerOutput(ctx, p)
	}

	// Sanitize event store file
	eventStoreFilePath := filepaths.EventStoreFilePath(e.Event.ExecutionID)
	err = sanitize.Instance.SanitizeFile(eventStoreFilePath)
	if err != nil {
		slog.Error("Failed to sanitize file", "eventStoreFilePath", eventStoreFilePath)
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(e.Event.ExecutionID)

	return nil
}
