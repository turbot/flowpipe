package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Load struct {
	RunID string `json:"run_id"`
}

type LoadHandler CommandHandler

func (h LoadHandler) HandlerName() string {
	return "command.load"
}

func (h LoadHandler) NewCommand() interface{} {
	return &Load{}
}

func (h LoadHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*Load)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.Loaded{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
