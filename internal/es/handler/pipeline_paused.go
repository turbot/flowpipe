package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelinePaused EventHandler

var pipelinePaused = event.PipelinePaused{}

func (h PipelinePaused) HandlerName() string {
	return pipelinePaused.HandlerName()
}

func (PipelinePaused) NewEvent() interface{} {
	return &event.PipelinePaused{}
}

func (h PipelinePaused) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)

	e, ok := ei.(*event.PipelinePaused)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePaused", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePaused")
	}

	logger.Info("[8] pipeline_paused event handler", "eventExecutionID", e.Event.ExecutionID)
	return nil
}
