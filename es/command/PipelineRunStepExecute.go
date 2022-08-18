package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunStepExecute struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Pipeline      pipeline.Pipeline      `json:"pipeline"`
	StepIndex     int                    `json:"step_index"`
}

type PipelineRunStepExecuteHandler CommandHandler

func (h PipelineRunStepExecuteHandler) HandlerName() string {
	return "pipeline.run.step_execute"
}

func (h PipelineRunStepExecuteHandler) NewCommand() interface{} {
	return &PipelineRunStepExecute{}
}

func (h PipelineRunStepExecuteHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunStepExecute)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), cmd)

	step := cmd.Pipeline.Steps[cmd.StepIndex]

	switch step.Type {
	case "http_request":
		{
			e := event.PipelineRunStepHTTPRequestPlanned{
				IdentityID:    cmd.IdentityID,
				WorkspaceID:   cmd.WorkspaceID,
				PipelineName:  cmd.PipelineName,
				PipelineInput: cmd.PipelineInput,
				RunID:         cmd.RunID,
				Timestamp:     time.Now(),
				Pipeline:      cmd.Pipeline,
				StepIndex:     cmd.StepIndex,
			}
			return h.EventBus.Publish(ctx, &e)
		}
	}

	e := event.PipelineRunFailed{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		PipelineInput: cmd.PipelineInput,
		RunID:         cmd.RunID,
		Timestamp:     time.Now(),
		ErrorMessage:  fmt.Sprintf("step_type_not_found: %s", step.Type),
	}
	return h.EventBus.Publish(ctx, &e)
}
