package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
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

	e := event.Queued{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
