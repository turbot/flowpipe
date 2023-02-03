package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStartHandler CommandHandler

func (h PipelineRunStartHandler) HandlerName() string {
	return "command.pipeline_run_start"
}

func (h PipelineRunStartHandler) NewCommand() interface{} {
	return &event.PipelineRunStart{}
}

func (h PipelineRunStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunStart)

	e := event.PipelineRunStarted{
		RunID:         cmd.RunID,
		SpanID:        cmd.SpanID,
		PipelineInput: cmd.PipelineInput,
		CreatedAt:     time.Now(),
		Pipeline:      cmd.Pipeline,
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
