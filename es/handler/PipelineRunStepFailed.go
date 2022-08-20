package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepFailed EventHandler

func (h PipelineRunStepFailed) HandlerName() string {
	return "pipeline.run.step_failed"
}

func (PipelineRunStepFailed) NewEvent() interface{} {
	return &event.PipelineRunStepFailed{}
}

func (h PipelineRunStepFailed) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineRunStepFailed)

	cmd := &command.PipelineRunFail{
		RunID:        e.RunID,
		ErrorMessage: e.ErrorMessage,
	}

	fmt.Printf("[handler] %s: %v\n", h.HandlerName(), cmd)

	return h.CommandBus.Send(ctx, cmd)
}
