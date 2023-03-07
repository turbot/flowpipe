package command

import (
	"context"

	"github.com/pkg/errors"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
	"github.com/turbot/steampipe-pipelines/primitive"
)

type PipelineStepStartHandler CommandHandler

func (h PipelineStepStartHandler) HandlerName() string {
	return "command.pipeline_step_start"
}

func (h PipelineStepStartHandler) NewCommand() interface{} {
	return &event.PipelineStepStart{}
}

func (h PipelineStepStartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelineStepStart)

	s, err := state.NewState(ctx, cmd.Event)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	// Load the pipeline definition
	// TODO - pipeline name needs to be read from the state
	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		e := event.Failed{
			Event:        event.NewFlowEvent(cmd.Event),
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
	case "pipeline":
		p := primitive.RunPipeline{}
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
			Event:        event.NewFlowEvent(cmd.Event),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	if step.Type == "pipeline" {
		e := event.PipelineStepStarted{
			Event:     event.NewFlowEvent(cmd.Event),
			StepIndex: cmd.StepIndex,
		}
		return h.EventBus.Publish(ctx, &e)
	}

	// All other primitives finish immediately.
	e := event.PipelineStepFinished{
		Event:     event.NewFlowEvent(cmd.Event),
		StepIndex: cmd.StepIndex,
		Output:    output,
	}

	return h.EventBus.Publish(ctx, &e)
}
