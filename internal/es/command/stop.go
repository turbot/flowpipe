package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type StopHandler CommandHandler

func (h StopHandler) HandlerName() string {
	return "command.stop"
}

func (h StopHandler) NewCommand() interface{} {
	return &event.Stop{}
}

func (h StopHandler) Handle(ctx context.Context, c interface{}) error {

	cmd, ok := c.(*event.Stop)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.Stop", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.Stop")
	}

	fplog.Logger(ctx).Info("(14) stop command handler", "executionID", cmd.Event.ExecutionID)

	e := event.Stopped{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
