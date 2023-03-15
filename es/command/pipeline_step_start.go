package command

import (
	"context"

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

	go func(ctx context.Context, c interface{}, h PipelineStepStartHandler) {

		cmd := c.(*event.PipelineStepStart)

		ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
		if err != nil {
			h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			return
		}

		defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
		if err != nil {
			h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			return
		}

		stepDefn := defn.Steps[cmd.StepName]

		var output *pipeline.Output

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
			h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			return
		}

		if err != nil {
			h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			return
		}

		if stepDefn.Type == "pipeline" {
			input := pipeline.PipelineInput{}
			if cmd.StepInput["input"] != nil {
				input = cmd.StepInput["input"].(pipeline.PipelineInput)
			}
			e, err := event.NewPipelineStepStarted(
				event.ForPipelineStepStart(cmd),
				event.WithNewChildPipelineExecutionID(),
				event.WithChildPipeline(cmd.StepInput["name"].(string), input))
			if err != nil {
				h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
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
			h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			return
		}

		h.EventBus.Publish(ctx, &e)

	}(ctx, c, h)

	return nil
}
