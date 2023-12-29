package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/output"
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
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
			evt.Event.ExecutionID, pipelineDefn.PipelineName, "finished")
		p.Output = evt.PipelineOutput
		output.RenderServerOutput(ctx, p)
	}

	if len(pipelineDefn.OutputConfig) > 0 {
		data[schema.BlockTypePipelineOutput] = evt.PipelineOutput
	}

	ex.Lock.Lock()
	defer ex.Lock.Unlock()

	err = ex.SaveToFile()
	if err != nil {
		slog.Error("pipeline_finished: Error saving execution", "error", err)
		// Should we raise pipeline fail here?
		return nil
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(evt.Event.ExecutionID)

	return nil
}
