package command

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
	"github.com/turbot/steampipe-pipelines/pipeline"
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

	go func() {

		cmd := c.(*event.PipelineStepStart)

		ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
		if err != nil {
			// TODO - should this return a failed event? how are errors caught here?
			e := event.PipelineFailed{
				Event:        event.NewFlowEvent(cmd.Event),
				ErrorMessage: err.Error(),
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
		if err != nil {
			e := event.PipelineFailed{
				Event:        event.NewFlowEvent(cmd.Event),
				ErrorMessage: err.Error(),
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		stepDefn := defn.Steps[cmd.StepName]

		var output pipeline.StepOutput

		switch stepDefn.Type {
		case "exec":
			p := primitive.Exec{}
			output, err = p.Run(ctx, cmd.StepInput)
		case "http_request":
			p := primitive.HTTPRequest{}
			output, err = p.Run(ctx, cmd.StepInput)
		case "pipeline":
			p := primitive.RunPipeline{}
			output, err = p.Run(ctx, cmd.StepInput)
		case "query":
			p := primitive.Query{}
			output, err = p.Run(ctx, cmd.StepInput)
		case "sleep":
			p := primitive.Sleep{}
			output, err = p.Run(ctx, cmd.StepInput)
		default:
			e := event.PipelineFailed{
				Event:        event.NewFlowEvent(cmd.Event),
				ErrorMessage: fmt.Sprintf("step type primitive not found: %s", stepDefn.Type),
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		if err != nil {
			e := event.Failed{
				Event:        event.NewFlowEvent(cmd.Event),
				ErrorMessage: err.Error(),
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		if stepDefn.Type == "pipeline" {
			input := pipeline.PipelineInput{}
			if stepDefn.Input["input"] != nil {
				input = stepDefn.Input["input"].(pipeline.PipelineInput)
			}
			e, err := event.NewPipelineStepStarted(
				event.ForPipelineStepStart(cmd),
				event.WithNewChildPipelineExecutionID(),
				event.WithChildPipeline(stepDefn.Input["name"].(string), input))
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(cmd.Event),
					ErrorMessage: err.Error(),
				}
				h.EventBus.Publish(ctx, &e)
				return
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		// All other primitives finish immediately.
		e, err := event.NewPipelineStepFinished(
			event.ForPipelineStepStartToPipelineStepFinished(cmd),
			event.WithStepOutput(output))
		if err != nil {
			e := event.PipelineFailed{
				Event:        event.NewFlowEvent(cmd.Event),
				ErrorMessage: err.Error(),
			}
			h.EventBus.Publish(ctx, &e)
			return
		}

		h.EventBus.Publish(ctx, &e)

	}()

	return nil
}
