package command

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type PipelineStepExecuteHandler CommandHandler

func (h PipelineStepExecuteHandler) HandlerName() string {
	return "command.pipeline_step_execute"
}

func (h PipelineStepExecuteHandler) NewCommand() interface{} {
	return &event.PipelineStepExecute{}
}

func (h PipelineStepExecuteHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelineStepExecute)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	s, err := state.NewState(ctx, cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// Load the pipeline definition
	// TODO - pipeline name needs to be read from the state
	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		e := event.Failed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now().UTC(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	step := defn.Steps[cmd.StepIndex]

	var output primitive.Output

	switch step.Type {
	case "exec":
		p := primitive.Exec{}
		output, err = p.Run(ctx, cmd.Input)
	case "http_request":
		p := primitive.HTTPRequest{}
		output, err = p.Run(ctx, cmd.Input)
	case "query":
		p := primitive.Query{}
		output, err = p.Run(ctx, cmd.Input)
	case "sleep":
		p := primitive.Sleep{}
		output, err = p.Run(ctx, cmd.Input)
	default:
		return errors.Errorf("step_type_not_found: %s", step.Type)
	}

	if err != nil {
		e := event.Failed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now().UTC(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineStepExecuted{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		StackID:   cmd.StackID,
		StepIndex: cmd.StepIndex,
		CreatedAt: time.Now().UTC(),
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
