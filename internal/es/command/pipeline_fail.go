package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
)

type PipelineFailHandler CommandHandler

func (h PipelineFailHandler) HandlerName() string {
	return "command.pipeline_fail"
}

func (h PipelineFailHandler) NewCommand() interface{} {
	return &event.PipelineFail{}
}

func (h PipelineFailHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineFail)
	return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFail(cmd)))
}
