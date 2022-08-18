package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunQueued EventHandler

func (h PipelineRunQueued) HandlerName() string {
	return "pipeline.run.queued"
}

func (PipelineRunQueued) NewEvent() interface{} {
	return &event.PipelineRunQueued{}
}

func (h PipelineRunQueued) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunQueued)

	cmd := &command.PipelineRunLoad{
		IdentityID:    e.IdentityID,
		WorkspaceID:   e.WorkspaceID,
		PipelineName:  e.PipelineName,
		PipelineInput: e.PipelineInput,
		RunID:         e.RunID,
	}

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

	return h.CommandBus.Send(ctx, cmd)
}
