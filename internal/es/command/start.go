package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type StartHandler CommandHandler

func (h StartHandler) HandlerName() string {
	return "command.start"
}

func (h StartHandler) NewCommand() interface{} {
	return &event.Start{}
}

func (h StartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd, ok := c.(*event.Start)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.Start", "actual", c)
		return pcerr.BadRequestWithMessage("invalid command type expected *event.Start")
	}

	fplog.Logger(ctx).Info("(13) start command handler", "executionID", cmd.Event.ExecutionID)

	e := event.Started{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
