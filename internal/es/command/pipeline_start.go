package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStartHandler CommandHandler

func (h PipelineStartHandler) HandlerName() string {
	return execution.PipelineStartCommand.HandlerName()
}

func (h PipelineStartHandler) NewCommand() interface{} {
	return &event.PipelineStart{}
}

func (h PipelineStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStart)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStart", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineStart")
	}

	e, err := event.NewPipelineStarted(event.ForPipelineStart(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStartToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
