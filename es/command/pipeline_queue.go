package command

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.PipelineQueued{
		Name:      cmd.Name,
		Input:     cmd.Input,
		RunID:     fmt.Sprintf("%s.%s", cmd.RunID, cmd.SpanID),
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
