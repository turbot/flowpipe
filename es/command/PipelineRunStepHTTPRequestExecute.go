package command

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunStepHTTPRequestExecute struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Pipeline      pipeline.Pipeline      `json:"pipeline"`
	StepIndex     int                    `json:"step_index"`
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
	if urli, ok := cmd.PipelineInput["url"]; ok {
		url = urli.(string)
	}
	if url == "" {
		e := event.PipelineRunFailed{
			IdentityID:   cmd.IdentityID,
			WorkspaceID:  cmd.WorkspaceID,
			PipelineName: cmd.PipelineName,
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
			IdentityID:    cmd.IdentityID,
			WorkspaceID:   cmd.WorkspaceID,
			PipelineName:  cmd.PipelineName,
			PipelineInput: cmd.PipelineInput,
			RunID:         cmd.RunID,
			Timestamp:     time.Now(),
			ErrorMessage:  err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e := event.PipelineRunFailed{
			IdentityID:    cmd.IdentityID,
			WorkspaceID:   cmd.WorkspaceID,
			PipelineName:  cmd.PipelineName,
			PipelineInput: cmd.PipelineInput,
			RunID:         cmd.RunID,
			Timestamp:     time.Now(),
			ErrorMessage:  err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineRunStepExecuted{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		PipelineInput: cmd.PipelineInput,
		RunID:         cmd.RunID,
		Timestamp:     time.Now(),
		Pipeline:      cmd.Pipeline,
		StepIndex:     cmd.StepIndex,
		Output:        string(body),
	}
	return h.EventBus.Publish(ctx, &e)
}
