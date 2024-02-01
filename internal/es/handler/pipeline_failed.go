package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/pipe-fittings/utils"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/types"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
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

	err := store.UpdatePipelineState(evt.Event.ExecutionID, "failed")
	if err != nil {
		slog.Error("pipeline_failed: Error updating pipeline state", "error", err)
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

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
		duration := utils.HumanizeDuration(evt.Event.CreatedAt.Sub(ex.PipelineExecutions[evt.PipelineExecutionID].StartTime))
		prefix := types.NewPrefixWithServer(pipelineDefn.PipelineName, types.NewServerOutputPrefixWithExecId(evt.Event.CreatedAt, "pipeline", &evt.Event.ExecutionID))
		pe := types.NewParsedEvent(prefix, evt.Event.ExecutionID, event.HandlerPipelineFailed, "", "")
		o := types.NewParsedErrorEvent(pe, evt.Errors, evt.PipelineOutput, &duration, false, true)
		output.RenderServerOutput(ctx, o)
	}

	err = ex.Save()
	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(evt.Event.ExecutionID)
	execution.CompletePipelineExecutionStepSemaphore(evt.PipelineExecutionID)
	err = execution.ReleasePipelineSemaphore(pipelineDefn)
	if err != nil {
		slog.Error("pipeline_finished: Error releasing pipeline semaphore", "error", err)
	}

	return nil
}
