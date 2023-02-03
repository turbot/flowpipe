package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepFailed EventHandler

func (h PipelineRunStepFailed) HandlerName() string {
	return "handler.pipeline_run_step_failed"
}

func (PipelineRunStepFailed) NewEvent() interface{} {
	return &event.PipelineRunStepFailed{}
}

func (h PipelineRunStepFailed) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepFailed)

	cmd := &event.PipelineRunFail{
		RunID:        e.RunID,
		SpanID:       e.SpanID,
		CreatedAt:    time.Now(),
		ErrorMessage: e.ErrorMessage,
	}

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

	return h.CommandBus.Send(ctx, cmd)
}
