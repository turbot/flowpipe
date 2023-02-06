package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type QueueHandler CommandHandler

func (h QueueHandler) HandlerName() string {
	return "command.queue"
}

func (h QueueHandler) NewCommand() interface{} {
	return &event.Queue{}
}

// Queue the mod for handling and execution
func (h QueueHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Queue)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.Queued{
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now().UTC(),
	}

	return h.EventBus.Publish(ctx, &e)
}
