package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStarted EventHandler

func (h PipelineRunStarted) HandlerName() string {
	return "pipeline.run.started"
}

func (PipelineRunStarted) NewEvent() interface{} {
	return &event.PipelineRunStarted{}
}

func (h PipelineRunStarted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStarted)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	/*
		// Mocking failure
		if rand.Float64() < 0.5 {
			e := event.PipelineRunFailed{
				IdentityID:   e.IdentityID,
				WorkspaceID:  e.WorkspaceID,
				PipelineName: e.PipelineName,
				RunID:        e.RunID,
				Timestamp:    time.Now(),
				ErrorMessage: "pipeline_run_failed",
			}
			fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)
			return h.EventBus.Publish(ctx, &e)
		}
	*/

	if len(e.Pipeline.Steps) <= 0 {
		// Nothing to do!
		cmd := &command.PipelineRunFinish{
			IdentityID:    e.IdentityID,
			WorkspaceID:   e.WorkspaceID,
			PipelineName:  e.PipelineName,
			PipelineInput: e.PipelineInput,
			RunID:         e.RunID,
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// Run the first step
	cmd := &command.PipelineRunStepExecute{
		IdentityID:    e.IdentityID,
		WorkspaceID:   e.WorkspaceID,
		PipelineName:  e.PipelineName,
		PipelineInput: e.PipelineInput,
		RunID:         e.RunID,
		Pipeline:      e.Pipeline,
		StepIndex:     0,
	}

	return h.CommandBus.Send(ctx, cmd)
}
