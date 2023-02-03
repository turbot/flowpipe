package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepExecuted EventHandler

func (h PipelineRunStepExecuted) HandlerName() string {
	return "handler.pipeline_run_step_executed"
}

func (PipelineRunStepExecuted) NewEvent() interface{} {
	return &event.PipelineRunStepExecuted{}
}

func (h PipelineRunStepExecuted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepExecuted)

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), e)

	if e.StepIndex >= len(e.Pipeline.Steps)-1 {
		// Nothing to do!
		cmd := &event.PipelineRunFinish{
			SpanID: e.SpanID,
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// We have another step to run
	cmd := &event.PipelineRunStepExecute{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
		StepID:    xid.New().String(),
		Pipeline:  e.Pipeline,
		StepIndex: e.StepIndex + 1,
	}

	return h.CommandBus.Send(ctx, cmd)
}
