package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
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

	e := event.Started{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
