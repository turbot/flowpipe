package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepExecuted EventHandler

func (h PipelineRunStepExecuted) HandlerName() string {
	return "pipeline.run.step_executed"
}

func (PipelineRunStepExecuted) NewEvent() interface{} {
	return &event.PipelineRunStepExecuted{}
}

func (h PipelineRunStepExecuted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepExecuted)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	if e.StepIndex >= len(e.Pipeline.Steps)-1 {
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

	// We have another step to run
	cmd := &command.PipelineRunStepExecute{
		IdentityID:    e.IdentityID,
		WorkspaceID:   e.WorkspaceID,
		PipelineName:  e.PipelineName,
		PipelineInput: e.PipelineInput,
		RunID:         e.RunID,
		Pipeline:      e.Pipeline,
		StepIndex:     e.StepIndex + 1,
	}

	return h.CommandBus.Send(ctx, cmd)
}
