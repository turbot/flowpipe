package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStarted EventHandler

func (h PipelineRunStarted) HandlerName() string {
	return "handler.pipeline_run_started"
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
				SpanID:        e.SpanID,
				Timestamp:    time.Now(),
				ErrorMessage: "pipeline_run_failed",
			}
			fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)
			return h.EventBus.Publish(ctx, &e)
		}
	*/

	if len(e.Pipeline.Steps) <= 0 {
		// Nothing to do!
		cmd := &event.PipelineRunFinish{
			RunID:     e.RunID,
			SpanID:    e.SpanID,
			CreatedAt: time.Now(),
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// Run the first step
	cmd := &event.PipelineRunStepExecute{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
		StepID:    xid.New().String(),
		Pipeline:  e.Pipeline,
		StepIndex: 0,
		StepInput: e.Pipeline.Steps[0].Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
