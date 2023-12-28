package handler

import (
	"context"
	"encoding/json"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
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

	e, ok := ei.(*event.PipelineFinished)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineFinished")
	}

	slog.Debug("pipeline_finished event handler", "executionID", e.Event.ExecutionID, "pipelineExecutionID", e.PipelineExecutionID)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}

	if parentStepExecution != nil {
		cmd, err := event.NewStepPipelineFinish(
			event.ForPipelineFinished(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID),

			// If StepForEach is not nil, it indicates that this pipeline execution is part of
			// for_each steps
			event.WithStepForEach(parentStepExecution.StepForEach))

		cmd.StepRetry = parentStepExecution.StepRetry
		cmd.StepInput = parentStepExecution.Input
		cmd.StepLoop = parentStepExecution.StepLoop

		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
		}

		return h.CommandBus.Send(ctx, cmd)

	}
	// Generate output data
	data, err := ex.PipelineData(e.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_finished (2)", "error", err)
	} else {
		jsonStr, _ := json.MarshalIndent(data, "", "  ")
		slog.Debug("json string", "json", string(jsonStr))
	}

	pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		slog.Error("Pipeline definition not found", "error", err)
		return err
	}

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutputPrefix(e.Event.CreatedAt, "pipeline"),
			e.Event.ExecutionID, pipelineDefn.PipelineName, "finished")
		p.Output = e.PipelineOutput
		output.RenderServerOutput(ctx, p)
	}

	if len(pipelineDefn.OutputConfig) > 0 {
		data[schema.BlockTypePipelineOutput] = e.PipelineOutput
	}

	eventStoreFilePath := filepaths.EventStoreFilePath(e.Event.ExecutionID)
	err = sanitize.Instance.SanitizeFile(eventStoreFilePath)
	if err != nil {
		slog.Error("Failed to sanitize file", "eventStoreFilePath", eventStoreFilePath)
	}

	// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
	event.ReleaseEventLogMutex(e.Event.ExecutionID)

	return nil
}
