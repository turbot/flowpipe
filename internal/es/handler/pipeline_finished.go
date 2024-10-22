package handler

import (
	"context"
	"log/slog"
	"slices"

	"github.com/turbot/pipe-fittings/utils"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/types"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return execution.PipelineFinishedEvent.HandlerName()
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.PipelineFinished)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineFinished")
	}

	slog.Debug("pipeline_finished event handler", "executionID", evt.Event.ExecutionID, "pipelineExecutionID", evt.PipelineExecutionID)

	err := store.UpdatePipelineState(evt.Event.ExecutionID, "finished")
	if err != nil {
		slog.Error("pipeline_finished: Error updating pipeline state", "error", err)
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_finished: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	parentStepExecution, err := ex.ParentStepExecution(evt.PipelineExecutionID)
	if err != nil {
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	if parentStepExecution != nil {
		cmd, err := event.NewStepPipelineFinish(
			event.ForPipelineFinished(evt),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID),

			// If StepForEach is not nil, it indicates that this pipeline execution is part of
			// for_each steps
			event.WithStepForEach(parentStepExecution.StepForEach))

		cmd.StepRetry = parentStepExecution.StepRetry
		cmd.StepInput = parentStepExecution.Input
		cmd.StepLoop = parentStepExecution.StepLoop

		if err != nil {
			err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
			return nil
		}

		return h.CommandBus.Send(ctx, cmd)

	}
	// Generate output data
	data, err := ex.PipelineData(evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_finished ", "error", err)
	}

	if output.IsServerMode {
		duration := utils.HumanizeDuration(evt.Event.CreatedAt.Sub(ex.PipelineExecutions[evt.PipelineExecutionID].StartTime))
		prefix := types.NewPrefixWithServer(pipelineDefn.PipelineName, types.NewServerOutputPrefixWithExecId(evt.Event.CreatedAt, "pipeline", &evt.Event.ExecutionID))
		pe := types.NewParsedEvent(prefix, evt.Event.ExecutionID, event.HandlerPipelineFinished, "", "")
		o := types.NewParsedEventWithOutput(pe, map[string]any{}, evt.PipelineOutput, &duration, true)
		output.RenderServerOutput(ctx, o)
	}

	if len(pipelineDefn.OutputConfig) > 0 {
		data[schema.BlockTypePipelineOutput] = evt.PipelineOutput
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	pipelineCompletionHandler(evt.Event.ExecutionID, evt.PipelineExecutionID, pipelineDefn, ex.PipelineExecutions[evt.PipelineExecutionID].StepExecutions)

	// raise execution plan command if this pipeline is in the root pipeline list
	if slices.Contains(ex.RootPipelines, evt.PipelineExecutionID) {
		cmd := event.ExecutionPlanFromPipelineFinished(evt)
		err = h.CommandBus.Send(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
	}

	return nil
}
