package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type PipelineStepFinishHandler CommandHandler

func (h PipelineStepFinishHandler) HandlerName() string {
	return "command.pipeline_step_finish"
}

func (h PipelineStepFinishHandler) NewCommand() interface{} {
	return &event.PipelineStepFinish{}
}

// There's only one use case for this, which is to handle the "Pipeline Step" finish command.
//
// Pipeline Step = step that launches another pipeline.
func (h PipelineStepFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStepFinish)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStepFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineStepFinish")
	}

	e, err := event.NewPipelineStepFinished(event.ForPipelineStepFinish(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepFinishToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
