package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineStartHandler CommandHandler

func (h PipelineStartHandler) HandlerName() string {
	return "command.pipeline_start"
}

func (h PipelineStartHandler) NewCommand() interface{} {
	return &event.PipelineStart{}
}

func (h PipelineStartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelineStart)

	e := event.PipelineStarted{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
