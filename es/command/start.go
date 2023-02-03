package command

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type StartHandler CommandHandler

func (h StartHandler) HandlerName() string {
	return "command.start"
}

func (h StartHandler) NewCommand() interface{} {
	return &event.Start{}
}

func (h StartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Start)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	// TODO - Start running handlers for the mod. After this, we should be
	// capturing and handling events.

	/*
		s, err := state.NewState(cmd.SpanID)
		if err != nil {
			// TODO - should this return a failed event? how are errors caught here?
			return err
		}
	*/

	e := event.Started{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		StackID:   xid.New().String(),
	}

	return h.EventBus.Publish(ctx, &e)
}
