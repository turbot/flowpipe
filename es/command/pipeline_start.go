package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineStartHandler CommandHandler

func (h PipelineStartHandler) HandlerName() string {
	return "command.pipeline_start"
}

func (h PipelineStartHandler) NewCommand() interface{} {
	return &event.PipelineStart{}
}

func (h PipelineStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineStart)
	e, err := event.NewPipelineStarted(event.ForPipelineStart(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStartToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
