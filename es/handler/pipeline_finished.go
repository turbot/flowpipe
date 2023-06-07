package handler

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return "handler.pipeline_finished"
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	e, ok := ei.(*event.PipelineFinished)
	if !ok {
		return fperr.BadRequestWithMessage("ei is not a PipelineFinished")
	}

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
		// Dump the final execution state
		_, err := json.MarshalIndent(ex, "", "  ")
		if err != nil {
			fplog.Logger(ctx).Error("pipeline_failed (1)", "error", err)
		}

		// Dump step outputs
		data, err := ex.PipelineData(e.PipelineExecutionID)
		if err != nil {
			fplog.Logger(ctx).Error("pipeline_failed (2)", "error", err)
		} else {
			jsonStr, _ := json.MarshalIndent(data, "", "  ")
			fplog.Logger(ctx).Info("json string", "json", string(jsonStr))
		}

		// Dump the snapshot
		snapshot, err := ex.Snapshot(e.PipelineExecutionID)
		if err != nil {
			fplog.Logger(ctx).Error("pipeline_failed (3)", "error", err)
		} else {
			jsonStr, _ := json.MarshalIndent(snapshot, "", "  ")

			filePath := path.Join(viper.GetString("output.dir"), e.Event.ExecutionID+".sps")
			_ = os.WriteFile(filePath, jsonStr, 0600)
		}

	}

	return nil
}
