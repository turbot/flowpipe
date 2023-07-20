package command

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PipelineCancelHandler CommandHandler

func (h PipelineCancelHandler) HandlerName() string {
	return "command.pipeline_cancel"
}

func (h PipelineCancelHandler) NewCommand() interface{} {
	return &event.PipelineCancel{}
}

func (h PipelineCancelHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)
	evt, ok := c.(*event.PipelineCancel)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineCancel", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineCancel")
	}

	e, err := event.NewPipelineCanceled(event.ForPipelineCancel(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineCancelToPipelineFailed(evt, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
