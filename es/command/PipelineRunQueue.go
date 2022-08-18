package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"

	"github.com/rs/xid"
)

type PipelineRunQueue struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
}

type PipelineRunQueueHandler CommandHandler

func (h PipelineRunQueueHandler) HandlerName() string {
	return "pipeline.run.queue"
}

func (h PipelineRunQueueHandler) NewCommand() interface{} {
	return &PipelineRunQueue{}
}

func (h PipelineRunQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunQueue)
	e := event.PipelineRunQueued{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		PipelineInput: cmd.PipelineInput,
		RunID:         xid.New().String(),
		Timestamp:     time.Now(),
	}
	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)
	return h.EventBus.Publish(ctx, &e)
}
