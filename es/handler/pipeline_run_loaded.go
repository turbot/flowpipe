package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunLoaded EventHandler

func (h PipelineRunLoaded) HandlerName() string {
	return "handler.pipeline_run_loaded"
}

func (PipelineRunLoaded) NewEvent() interface{} {
	return &event.PipelineRunLoaded{}
}

func (h PipelineRunLoaded) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunLoaded)

	cmd := &command.PipelineRunStart{
		IdentityID:    e.IdentityID,
		WorkspaceID:   e.WorkspaceID,
		PipelineName:  e.PipelineName,
		PipelineInput: e.PipelineInput,
		RunID:         e.RunID,
		Pipeline:      e.Pipeline,
	}

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

	return h.CommandBus.Send(ctx, cmd)
}
