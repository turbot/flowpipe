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

	// TODO - We should do actual loading of the mod at this point.
	// In particular, we need to read in any triggers that are being
	// handled by the mod. These loaded triggers will be used until
	// the mod is reloaded.

	e := event.Loaded{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
