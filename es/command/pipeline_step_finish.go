package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStepFinishHandler CommandHandler

func (h PipelineStepFinishHandler) HandlerName() string {
	return "command.pipeline_step_finish"
}

func (h PipelineStepFinishHandler) NewCommand() interface{} {
	return &event.PipelineStepFinish{}
}

func (h PipelineStepFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineStepFinish)
	e, err := event.NewPipelineStepFinished(event.ForPipelineStepFinish(cmd))
	if err != nil {
		return err
	}
	return h.EventBus.Publish(ctx, &e)
}
