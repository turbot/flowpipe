package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type QueueHandler CommandHandler

func (h QueueHandler) HandlerName() string {
	return "command.queue"
}

func (h QueueHandler) NewCommand() interface{} {
	return &event.Queue{}
}

func (h QueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.Queue)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.Queue", "actual", c)
		return pcerr.BadRequestWithMessage("invalid command type expected *event.Queue")
	}

	e := event.Queued{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
