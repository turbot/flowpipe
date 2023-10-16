package handler

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/schema"
)

type PipelineFailed EventHandler

func (h PipelineFailed) HandlerName() string {
	return "handler.pipeline_failed"
}

func (PipelineFailed) NewEvent() interface{} {
	return &event.PipelineFailed{}
}

func (h PipelineFailed) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineFailed)

	logger := fplog.Logger(ctx)
	logger.Error("pipeline_failed handler", "event", e)

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
		cmd, err := event.NewPipelineStepFinish(
			event.ForPipelineFailed(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID))

		if err != nil {
			logger.Error("pipeline_failed error creating pipeline step finish event", "error", err)
			return err
		}

		return h.CommandBus.Send(ctx, &cmd)
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

		if len(pipelineDefn.OutputConfig) > 0 || (e.PipelineOutput != nil && e.PipelineOutput["errors"] != nil) {
			data[schema.BlockTypePipelineOutput] = e.PipelineOutput
		}

		// Dump the output
		jsonStr, _ := json.MarshalIndent(data, "", "  ")
		filePath := path.Join(viper.GetString(constants.ArgOutputDir), e.Event.ExecutionID+"_output.json")
		_ = os.WriteFile(filePath, jsonStr, 0600)

		snapshot, err := ex.Snapshot(e.PipelineExecutionID)
		if err != nil {
			logger.Error("pipeline_failed error generating snapshot", "error", err)
		} else {
			jsonStr, err := json.MarshalIndent(snapshot, "", "  ")

			if err != nil {
				logger.Error("pipeline_failed error generating snapshot", "error", err)
				return err
			}

			filePath := path.Join(viper.GetString(constants.ArgOutputDir), e.Event.ExecutionID+".sps")
			_ = os.WriteFile(filePath, jsonStr, 0600)
		}
	}

	return nil
}
