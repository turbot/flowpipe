package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineCancelHandler CommandHandler

func (h PipelineCancelHandler) HandlerName() string {
	return "command.pipeline_cancel"
}

func (h PipelineCancelHandler) NewCommand() interface{} {
	return &event.PipelineCancel{}
}

func (h PipelineCancelHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineCancel)

	e, err := event.NewPipelineCanceled(event.ForPipelineCancel(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineCancelToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
