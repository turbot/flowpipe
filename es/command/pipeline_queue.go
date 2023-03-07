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

	e := event.PipelineQueued{
		Event: event.NewFlowEvent(cmd.Event),
		Name:  cmd.Name,
		Input: cmd.Input,
	}

	return h.EventBus.Publish(ctx, &e)
}
