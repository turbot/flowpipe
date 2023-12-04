package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineQueueHandler CommandHandler

func (h PipelineQueueHandler) HandlerName() string {
	return execution.PipelineQueueCommand.HandlerName()
}

func (h PipelineQueueHandler) NewCommand() interface{} {
	return &event.PipelineQueue{}
}

func (h PipelineQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineQueue)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineQueue")
	}

	e, err := event.NewPipelineQueued(event.ForPipelineQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
