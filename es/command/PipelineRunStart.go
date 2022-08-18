package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunStart struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Pipeline      pipeline.Pipeline      `json:"pipeline"`
}

type PipelineRunStartHandler CommandHandler

func (h PipelineRunStartHandler) HandlerName() string {
	return "pipeline.run.start"
}

func (h PipelineRunStartHandler) NewCommand() interface{} {
	return &PipelineRunStart{}
}

func (h PipelineRunStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunStart)

	e := event.PipelineRunStarted{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		RunID:         cmd.RunID,
		PipelineInput: cmd.PipelineInput,
		Timestamp:     time.Now(),
		Pipeline:      cmd.Pipeline,
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
