package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFailed EventHandler

func (h PipelineRunFailed) HandlerName() string {
	return "handler.pipeline_run_failed"
}

func (PipelineRunFailed) NewEvent() interface{} {
	return &event.PipelineRunFailed{}
}

func (h PipelineRunFailed) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunFailed)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	return nil

	/*
		cmd := &command.PipelineRunLoad{
			IdentityID:   e.IdentityID,
			WorkspaceID:  e.WorkspaceID,
			PipelineName: e.PipelineName,
			RunID:        e.RunID,
		}

		fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

		return h.CommandBus.Send(ctx, cmd)
	*/
}
