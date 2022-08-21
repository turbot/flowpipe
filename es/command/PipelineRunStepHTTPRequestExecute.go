package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type PipelineRunStepHTTPRequestExecute struct {
	RunID string                 `json:"run_id"`
	Input map[string]interface{} `json:"input"`
}

type PipelineRunStepHTTPRequestExecuteHandler CommandHandler

func (h PipelineRunStepHTTPRequestExecuteHandler) HandlerName() string {
	return "pipeline.run.step_http_request_execute"
}

func (h PipelineRunStepHTTPRequestExecuteHandler) NewCommand() interface{} {
	return &PipelineRunStepHTTPRequestExecute{}
}

func (h PipelineRunStepHTTPRequestExecuteHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunStepHTTPRequestExecute)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), cmd)

	hr := primitive.HTTPRequest{}

	output, err := hr.Run(ctx, cmd.Input)
	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			Timestamp:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineRunStepExecuted{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
