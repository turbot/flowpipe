package command

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PipelineStepFinishHandler CommandHandler

func (h PipelineStepFinishHandler) HandlerName() string {
	return "command.pipeline_step_finish"
}

func (h PipelineStepFinishHandler) NewCommand() interface{} {
	return &event.PipelineStepFinish{}
}

func (h PipelineStepFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStepFinish)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStepFinish", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineStepFinish")
	}

	fplog.Logger(ctx).Info("(11) pipeline_step_finish command handler", "executionID", cmd.Event.ExecutionID, "cmd", cmd)

	e, err := event.NewPipelineStepFinished(event.ForPipelineStepFinish(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
