package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineStartHandler CommandHandler

func (h PipelineStartHandler) HandlerName() string {
	return "command.pipeline_start"
}

func (h PipelineStartHandler) NewCommand() interface{} {
	return &event.PipelineStart{}
}

func (h PipelineStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStart)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStart", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineStart")
	}

	fplog.Logger(ctx).Info("(10) pipeline_start command handler", "executionID", cmd.Event.ExecutionID)

	e, err := event.NewPipelineStarted(event.ForPipelineStart(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStartToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
