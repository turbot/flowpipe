package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFinish struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
}

type PipelineRunFinishHandler CommandHandler

func (h PipelineRunFinishHandler) HandlerName() string {
	return "pipeline.run.finish"
}

func (h PipelineRunFinishHandler) NewCommand() interface{} {
	return &PipelineRunFinish{}
}

func (h PipelineRunFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunFinish)

	e := event.PipelineRunFinished{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		PipelineInput: cmd.PipelineInput,
		RunID:         cmd.RunID,
		Timestamp:     time.Now(),
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
