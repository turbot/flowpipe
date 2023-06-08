package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
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
	cmd, ok := c.(*event.PipelineCancel)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineCancel", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineCancel")
	}

	logger.Info("(2) pipeline_cancel command handler")

	e, err := event.NewPipelineCanceled(event.ForPipelineCancel(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineCancelToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
