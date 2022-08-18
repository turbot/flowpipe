package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepHTTPRequestPlanned EventHandler

func (h PipelineRunStepHTTPRequestPlanned) HandlerName() string {
	return "pipeline.run.step_http_request_planned"
}

func (PipelineRunStepHTTPRequestPlanned) NewEvent() interface{} {
	return &event.PipelineRunStepHTTPRequestPlanned{}
}

func (h PipelineRunStepHTTPRequestPlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepHTTPRequestPlanned)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	// We have another step to run
	cmd := &command.PipelineRunStepHTTPRequestExecute{
		IdentityID:    e.IdentityID,
		WorkspaceID:   e.WorkspaceID,
		PipelineName:  e.PipelineName,
		PipelineInput: e.PipelineInput,
		RunID:         e.RunID,
		Pipeline:      e.Pipeline,
		StepIndex:     e.StepIndex,
	}

	return h.CommandBus.Send(ctx, cmd)
}
