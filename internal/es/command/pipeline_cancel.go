package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineCancelHandler CommandHandler

var pipelineCancel = event.PipelineCancel{}

func (h PipelineCancelHandler) HandlerName() string {
	return pipelineCancel.HandlerName()
}

func (h PipelineCancelHandler) NewCommand() interface{} {
	return &event.PipelineCancel{}
}

func (h PipelineCancelHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)
	evt, ok := c.(*event.PipelineCancel)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineCancel", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineCancel")
	}

	e, err := event.NewPipelineCanceled(event.ForPipelineCancel(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineCancelToPipelineFailed(evt, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
