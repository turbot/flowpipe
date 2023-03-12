package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineQueueHandler CommandHandler

func (h PipelineQueueHandler) HandlerName() string {
	return "command.pipeline_queue"
}

func (h PipelineQueueHandler) NewCommand() interface{} {
	return &event.PipelineQueue{}
}

func (h PipelineQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineQueue)
	e, err := event.NewPipelineQueued(event.ForPipelineQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
