package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type PipelineRunStepHTTPRequestExecuteHandler CommandHandler

func (h PipelineRunStepHTTPRequestExecuteHandler) HandlerName() string {
	return "command.pipeline_run_step_execute_httprequest"
}

func (h PipelineRunStepHTTPRequestExecuteHandler) NewCommand() interface{} {
	return &event.PipelineRunStepHTTPRequestExecute{}
}

func (h PipelineRunStepHTTPRequestExecuteHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunStepHTTPRequestExecute)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), cmd)

	hr := primitive.HTTPRequest{}

	output, err := hr.Run(ctx, cmd.Input)
	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineRunStepExecuted{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
