package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Stop struct {
	RunID string `json:"run_id"`
}

type StopHandler CommandHandler

func (h StopHandler) HandlerName() string {
	return "command.stop"
}

func (h StopHandler) NewCommand() interface{} {
	return &Stop{}
}

func (h StopHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*Stop)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	e := event.Stopped{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
		// TODO - Output
	}

	return h.EventBus.Publish(ctx, &e)
}
