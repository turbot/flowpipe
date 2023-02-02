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

func (h QueueHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Queue)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.Queued{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
