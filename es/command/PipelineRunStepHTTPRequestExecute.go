package command

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
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

	var url string
	if urli, ok := cmd.Input["url"]; ok {
		url = urli.(string)
	}
	if url == "" {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			Timestamp:    time.Now(),
			ErrorMessage: "http_request requires url input",
		}
		return h.EventBus.Publish(ctx, &e)
	}

	//step := cmd.Pipeline.Steps[cmd.StepIndex]

	resp, err := http.Get(url)
	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			Timestamp:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
		Output:    map[string]interface{}{"body": string(body)},
	}
	return h.EventBus.Publish(ctx, &e)
}
