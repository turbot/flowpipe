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

	snapshot, err := ex.Snapshot(e.PipelineExecutionID)
	if err != nil {
		logger.Error("pipeline_failed error generating snapshot", "error", err)
	} else {
		jsonStr, err := json.MarshalIndent(snapshot, "", "  ")

		if err != nil {
			logger.Error("pipeline_failed error generating snapshot", "error", err)
			return err
		}

		filePath := path.Join(viper.GetString("output.dir"), e.Event.ExecutionID+".sps")
		_ = os.WriteFile(filePath, jsonStr, 0600)
	}

	return nil
}
