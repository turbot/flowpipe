package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
)

type LoadHandler CommandHandler

func (h LoadHandler) HandlerName() string {
	return "command.load"
}

func (h LoadHandler) NewCommand() interface{} {
	return &event.Load{}
}

func (h LoadHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Load)

	// TODO - We should do actual loading of the mod at this point.  In
	// particular, we need to read in any triggers that are being handled by the
	// mod. These loaded triggers will be used until the mod is reloaded.

	e := event.Loaded{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
