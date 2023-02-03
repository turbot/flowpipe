package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type StopHandler CommandHandler

func (h StopHandler) HandlerName() string {
	return "command.stop"
}

func (h StopHandler) NewCommand() interface{} {
	return &event.Stop{}
}

func (h StopHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Stop)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	e := event.Stopped{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		// TODO - Output
	}

	return h.EventBus.Publish(ctx, &e)
}
