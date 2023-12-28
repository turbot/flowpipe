package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/pipe-fittings/modconfig"
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

	var parentStepExecution *execution.StepExecution
	var pipelineDefn *modconfig.Pipeline
	var ex *execution.ExecutionInMemory
	var err error

	executionID := evt.Event.ExecutionID

	if execution.ExecutionMode == "in-memory" {
		ex, pipelineDefn, err = execution.GetPipelineDefnFromExecution(executionID, evt.PipelineExecutionID)
		if err != nil {
			slog.Error("pipeline_finished: Error loading pipeline execution", "error", err)
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		}

		parentStepExecution, err = ex.ParentStepExecution(evt.PipelineExecutionID)
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		}

	} else {
		ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		}

		parentStepExecution, err = ex.ParentStepExecution(evt.PipelineExecutionID)
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		}

		pipelineDefn, err = ex.PipelineDefinition(evt.PipelineExecutionID)
		if err != nil {
			slog.Error("Pipeline definition not found", "error", err)
			return err
		}
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
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(evt, err)))
		}

		return h.CommandBus.Send(ctx, cmd)

	}
	// Generate output data
	data, err := ex.PipelineData(evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_finished (2)", "error", err)
	} else {
		jsonStr, _ := json.MarshalIndent(data, "", "  ")
		slog.Debug("json string", "json", string(jsonStr))
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

	eventStoreFilePath := filepaths.EventStoreFilePath(evt.Event.ExecutionID)
	err = sanitize.Instance.SanitizeFile(eventStoreFilePath)
	if err != nil {
		slog.Error("Failed to sanitize file", "eventStoreFilePath", eventStoreFilePath)
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(evt.Event.ExecutionID)

	return nil
}
