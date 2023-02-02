package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFinished EventHandler

func (h PipelineRunFinished) HandlerName() string {
	return "handler.pipeline_run_finished"
}

func (PipelineRunFinished) NewEvent() interface{} {
	return &event.PipelineRunFinished{}
}

func (h PipelineRunFinished) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunFinished)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	return nil

	/*
		cmd := &command.PipelineRunFinish{
			IdentityID:   e.IdentityID,
			WorkspaceID:  e.WorkspaceID,
			PipelineName: e.PipelineName,
			RunID:        e.RunID,
		}

		fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

		return h.CommandBus.Send(ctx, cmd)
	*/
}
