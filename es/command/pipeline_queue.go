package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineQueueHandler CommandHandler

func (h PipelineQueueHandler) HandlerName() string {
	return "command.pipeline_queue"
}

func (h PipelineQueueHandler) NewCommand() interface{} {
	return &event.PipelineQueue{}
}

func (h PipelineQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineQueue)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineQueue", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineQueue")
	}

	fplog.Logger(ctx).Info("(8) pipeline_queue command handler", "executionID", cmd.Event.ExecutionID)

	e, err := event.NewPipelineQueued(event.ForPipelineQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
