package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFailHandler CommandHandler

func (h PipelineRunFailHandler) HandlerName() string {
	return "command.pipeline_run_fail"
}

func (h PipelineRunFailHandler) NewCommand() interface{} {
	return &event.PipelineRunFail{}
}

func (h PipelineRunFailHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunFail)

	e := event.PipelineRunFailed{
		RunID:        cmd.RunID,
		SpanID:       cmd.SpanID,
		CreatedAt:    time.Now(),
		ErrorMessage: cmd.ErrorMessage,
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
