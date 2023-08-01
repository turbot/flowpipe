package handler

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return "handler.pipeline_finished"
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelineFinished)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelineFinished", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.PipelineFinished")
	}

	logger.Info("pipeline_finished event handler", "executionID", e.Event.ExecutionID, "pipelineExecutionID", e.PipelineExecutionID)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}

	if parentStepExecution != nil {
		cmd, err := event.NewPipelineStepFinish(
			event.ForPipelineFinished(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, &cmd)
	} else {
		// Generate output data
		data, err := ex.PipelineData(e.PipelineExecutionID)
		if err != nil {
			logger.Error("pipeline_finished (2)", "error", err)
		} else {
			jsonStr, _ := json.MarshalIndent(data, "", "  ")
			logger.Debug("json string", "json", string(jsonStr))
		}

		pipelineDefn, err := ex.PipelineDefinition(e.PipelineExecutionID)
		if err != nil {
			logger.Error("Pipeline definition not found", "error", err)
			return err
		}

		if len(pipelineDefn.Outputs) > 0 {
			data[schema.BlockTypePipelineOutput] = e.PipelineOutput
		}

		// Dump the output
		jsonStr, _ := json.MarshalIndent(data, "", "  ")
		filePath := path.Join(viper.GetString("output.dir"), e.Event.ExecutionID+"_output.json")
		_ = os.WriteFile(filePath, jsonStr, 0600)

		// Dump the snapshot
		snapshot, err := ex.Snapshot(e.PipelineExecutionID)
		if err != nil {
			logger.Error("pipeline_finished (3)", "error", err)
			return err
		}

		jsonStr, _ = json.MarshalIndent(snapshot, "", "  ")
		filePath = path.Join(viper.GetString("output.dir"), e.Event.ExecutionID+".sps")
		_ = os.WriteFile(filePath, jsonStr, 0600)

	}

	return nil
}
