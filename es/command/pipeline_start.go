package command

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.PipelineStarted{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		StackID:   cmd.StackID,
		CreatedAt: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
