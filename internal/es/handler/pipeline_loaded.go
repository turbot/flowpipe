package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type PipelineLoaded EventHandler

func (h PipelineLoaded) HandlerName() string {
	return "handler.pipeline_loaded"
}

func (PipelineLoaded) NewEvent() interface{} {
	return &event.PipelineLoaded{}
}

func (h PipelineLoaded) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineLoaded)

	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return pcerr.BadRequestWithMessage("invalid event type expected *event.PipelineLoaded")
	}

	cmd, err := event.NewPipelineStart(event.ForPipelineLoaded(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineLoadedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
