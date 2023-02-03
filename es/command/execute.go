package command

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type ExecuteHandler CommandHandler

func (h ExecuteHandler) HandlerName() string {
	return "command.execute"
}

func (h ExecuteHandler) NewCommand() interface{} {
	return &event.Execute{}
}

func (h ExecuteHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Execute)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	/*
		s, err := state.NewState(cmd.SpanID)
		if err != nil {
			// TODO - should this return a failed event? how are errors caught here?
			return err
		}
	*/

	// Load the pipeline definition
	// TODO - pipeline name needs to be read from the state
	//defn, err := PipelineDefinition(s.PipelineName)
	defn, err := PipelineDefinition("my_pipeline_0")
	if err != nil {
		e := event.Failed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
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
	default:
		return errors.Errorf("step_type_not_found: %s", step.Type)
	}

	if err != nil {
		e := event.Failed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.Executed{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		StackID:   cmd.StackID,
		CreatedAt: time.Now(),
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
