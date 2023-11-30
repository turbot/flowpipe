package handler

import (
	"context"
	"encoding/json"
	"os"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/schema"
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

	logger := fplog.Logger(ctx)
	logger.Debug("pipeline_failed handler", "event", e)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		logger.Error("pipeline_failed error constructing execution", "error", err)
		return err
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		// We're already in a pipeline failed event handler
		logger.Error("pipeline_failed error getting parent step execution", "error", err)
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
			logger.Error("pipeline_failed error creating pipeline step finish event", "error", err)
			return err
		}

		return h.CommandBus.Send(ctx, cmd)
	} else {
		// Generate output data
		data, err := ex.PipelineData(e.PipelineExecutionID)
		if err != nil {
			logger.Error("pipeline_failed", "error", err)
		} else {
			jsonStr, _ := json.MarshalIndent(data, "", "  ")
			logger.Debug("json string", "json", string(jsonStr))
		}

		pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
		if err != nil {
			logger.Error("Pipeline definition not found", "error", err)
			return err
		}

		pipelineErrors := e.Errors
		if len(pipelineErrors) > 0 {
			if e.PipelineOutput == nil {
				e.PipelineOutput = map[string]interface{}{}
			}
			e.PipelineOutput["errors"] = pipelineErrors
		}

		if len(pipelineDefn.OutputConfig) > 0 || (e.PipelineOutput != nil && e.PipelineOutput["errors"] != nil) {
			data[schema.BlockTypePipelineOutput] = e.PipelineOutput
		}

		// Dump the output
		jsonStr, _ := json.MarshalIndent(data, "", "  ")
		outputPath := filepaths.OutputFilePath(e.Event.ExecutionID)
		_ = os.WriteFile(outputPath, jsonStr, 0600)

		snapshot, err := ex.Snapshot(e.PipelineExecutionID)
		if err != nil {
			logger.Error("pipeline_failed error generating snapshot", "error", err)
		} else {
			jsonStr, err := json.MarshalIndent(snapshot, "", "  ")

			if err != nil {
				logger.Error("pipeline_failed error generating snapshot", "error", err)
				return err
			}

			snapshotPath := filepaths.SnapshotFilePath(e.Event.ExecutionID)
			_ = os.WriteFile(snapshotPath, jsonStr, 0600)
		}
		// release the execution mutex (do the same thing for pipeline_failed and pipeline_finished)
		event.ReleaseEventLogMutex(e.Event.ExecutionID)
	}

	return nil
}
