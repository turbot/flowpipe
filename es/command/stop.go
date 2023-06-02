package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
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

	e := event.Stopped{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
