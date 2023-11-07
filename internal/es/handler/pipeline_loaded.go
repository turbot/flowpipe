package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineLoaded EventHandler

var pipelineLoaded = event.PipelineLoaded{}

func (h PipelineLoaded) HandlerName() string {
	return pipelineLoaded.HandlerName()
}

func (PipelineLoaded) NewEvent() interface{} {
	return &event.PipelineLoaded{}
}

func (h PipelineLoaded) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.PipelineLoaded)

	if !ok {
		logger.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineLoaded")
	}

	cmd, err := event.NewPipelineStart(event.ForPipelineLoaded(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineLoadedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
