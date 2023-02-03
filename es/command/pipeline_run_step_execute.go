package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunStepExecuteHandler CommandHandler

func (h PipelineRunStepExecuteHandler) HandlerName() string {
	return "command.pipeline_run_step_execute"
}

func (h PipelineRunStepExecuteHandler) NewCommand() interface{} {
	return &event.PipelineRunStepExecute{}
}

func (h PipelineRunStepExecuteHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunStepExecute)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), cmd)

	step := cmd.Pipeline.Steps[cmd.StepIndex]

	switch step.Type {
	case "exec", "http_request":
		{
			e := event.PipelineRunStepPrimitivePlanned{
				RunID:     cmd.RunID,
				SpanID:    cmd.SpanID,
				CreatedAt: time.Now(),
				StepID:    cmd.StepID,
				Primitive: step.Type,
				Input:     cmd.StepInput,
			}
			return h.EventBus.Publish(ctx, &e)
		}
		/*
			case "http_request":
				{
					e := event.PipelineRunStepHTTPRequestPlanned{
						SpanID:     cmd.SpanID,
						Timestamp: time.Now(),
						StepID:    cmd.StepID,
						Input:     cmd.StepInput,
					}
					return h.EventBus.Publish(ctx, &e)
				}
		*/
	}

	// Need StepID in the failed status
	e := event.PipelineRunFailed{
		RunID:        cmd.RunID,
		SpanID:       cmd.SpanID,
		CreatedAt:    time.Now(),
		ErrorMessage: fmt.Sprintf("step_type_not_found: %s", step.Type),
	}
	return h.EventBus.Publish(ctx, &e)
}
