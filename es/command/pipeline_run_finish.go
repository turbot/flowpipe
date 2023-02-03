package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFinishHandler CommandHandler

func (h PipelineRunFinishHandler) HandlerName() string {
	return "command.pipeline_run_finish"
}

func (h PipelineRunFinishHandler) NewCommand() interface{} {
	return &event.PipelineRunFinish{}
}

func (h PipelineRunFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunFinish)

	e := event.PipelineRunFinished{
		Name:      cmd.Name,
		Input:     cmd.Input,
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
