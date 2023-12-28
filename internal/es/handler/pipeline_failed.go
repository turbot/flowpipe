package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"

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
	evt := ei.(*event.PipelineFailed)

	slog.Debug("pipeline_failed handler", "event", evt)

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_failed error loading pipeline execution", "error", err)
		return err
	}

	parentStepExecution, err := ex.ParentStepExecution(evt.PipelineExecutionID)
	if err != nil {
		// We're already in a pipeline failed event handler
		slog.Error("pipeline_failed error getting parent step execution", "error", err)
		return err
	}

	if parentStepExecution != nil {
		cmd, err := event.NewStepPipelineFinish(
			event.ForPipelineFailed(evt),
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

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
			evt.Event.ExecutionID, pipelineDefn.PipelineName, "failed")
		p.Errors = evt.Errors
		p.Output = evt.PipelineOutput
		output.RenderServerOutput(ctx, p)
	}

	ex.Lock.Lock()
	defer ex.Lock.Unlock()

	err = ex.SaveToFile()
	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	// Sanitize event store file
	eventStoreFilePath := filepaths.EventStoreFilePath(evt.Event.ExecutionID)
	err = sanitize.Instance.SanitizeFile(eventStoreFilePath)
	if err != nil {
		slog.Error("Failed to sanitize file", "eventStoreFilePath", eventStoreFilePath)
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(evt.Event.ExecutionID)

	return nil
}
