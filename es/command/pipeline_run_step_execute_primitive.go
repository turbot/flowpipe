package command

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type PipelineRunStepPrimitiveExecuteHandler CommandHandler

func (h PipelineRunStepPrimitiveExecuteHandler) HandlerName() string {
	return "command.pipeline_run_step_execute_primitive"
}

func (h PipelineRunStepPrimitiveExecuteHandler) NewCommand() interface{} {
	return &event.PipelineRunStepPrimitiveExecute{}
}

func (h PipelineRunStepPrimitiveExecuteHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineRunStepPrimitiveExecute)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), cmd)

	var output primitive.Output
	var err error

	switch cmd.Primitive {
	case "exec":
		p := primitive.Exec{}
		output, err = p.Run(ctx, cmd.Input)
	case "http_request":
		p := primitive.HTTPRequest{}
		output, err = p.Run(ctx, cmd.Input)
	default:
		return errors.Errorf("primitive_not_found: %s", cmd.Primitive)
	}

	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineRunStepExecuted{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
