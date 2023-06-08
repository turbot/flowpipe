package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/primitive"
	"github.com/turbot/flowpipe/types"
)

type PipelineStepStartHandler CommandHandler

func (h PipelineStepStartHandler) HandlerName() string {
	return "command.pipeline_step_start"
}

func (h PipelineStepStartHandler) NewCommand() interface{} {
	return &event.PipelineStepStart{}
}

// * This is the handler that will actually execute the primitive
func (h PipelineStepStartHandler) Handle(ctx context.Context, c interface{}) error {

	go func(ctx context.Context, c interface{}, h PipelineStepStartHandler) {

		cmd, ok := c.(*event.PipelineStepStart)
		if !ok {
			fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStepStart", "actual", c)
			return
		}

		fplog.Logger(ctx).Info("(11) pipeline_step_start command handler", "executionID", cmd.Event.ExecutionID)

		ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err2)
			}
			return
		}

		defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err2)
			}
			return
		}

		stepDefn := defn.Steps[cmd.StepName]

		var output *types.Output

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
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err2)
			}
			return
		}

		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err2)
			}
			return
		}

		if stepDefn.Type == "pipeline" {
			args := types.Input{}
			if cmd.StepInput["args"] != nil {
				args = cmd.StepInput["args"].(map[string]interface{})
			}
			e, err := event.NewPipelineStepStarted(
				event.ForPipelineStepStart(cmd),
				event.WithNewChildPipelineExecutionID(),
				event.WithChildPipeline(cmd.StepInput["name"].(string), args))
			if err != nil {
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					fplog.Logger(ctx).Error("Error publishing event", "error", err2)
				}
				return
			}
			err = h.EventBus.Publish(ctx, &e)
			if err != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err)
			}
			return
		}

		// All other primitives finish immediately.
		e, err := event.NewPipelineStepFinished(
			event.ForPipelineStepStartToPipelineStepFinished(cmd),
			event.WithStepOutput(output))
		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				fplog.Logger(ctx).Error("Error publishing event", "error", err2)
			}
			return
		}

		err = h.EventBus.Publish(ctx, &e)
		if err != nil {
			fplog.Logger(ctx).Error("Error publishing event", "error", err)
		}

	}(ctx, c, h)

	return nil
}
