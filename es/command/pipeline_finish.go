package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return "command.pipeline_finish"
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelineFinish)

	e := event.PipelineFinished{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
