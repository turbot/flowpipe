package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/schema"
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

		logger := fplog.Logger(ctx)

		cmd, ok := c.(*event.PipelineStepStart)
		if !ok {
			logger.Error("invalid command type", "expected", "*event.PipelineStepStart", "actual", c)
			return
		}

		ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
		if err != nil {
			logger.Error("Error loading pipeline execution", "error", err)

			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
		if err != nil {
			logger.Error("Error loading pipeline definition", "error", err)

			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		stepDefn := pipelineDefn.GetStep(cmd.StepName)

		// Check if the step is an "if" condition
		if stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf] != nil {
			var err error
			expr := stepDefn.GetUnresolvedAttributes()[schema.AttributeTypeIf]

			// Evaluate the expression
			evalContext, err := ex.BuildEvalContext(pipelineDefn)
			if err != nil {
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return
			}

			val, diags := expr.Value(evalContext)
			if len(diags) > 0 {
				err = pipeparser.DiagsToError("diags", diags)

				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return
			}

			if val.False() {
				logger.Info("if condition not met for step", "step", stepDefn.GetName())

				output := &types.StepOutput{}

				endStep(cmd, output, logger, h, ctx)
				return
			}
		}

		var output *types.StepOutput
		var primitiveError error
		switch stepDefn.GetType() {
		case schema.BlockTypePipelineExec:
			p := primitive.Exec{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepHttp:
			p := primitive.HTTPRequest{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case "pipeline":
			p := primitive.RunPipeline{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineQuery:
			p := primitive.Query{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepSleep:
			p := primitive.Sleep{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		// TODO: remove this debug primitive (?)
		case schema.BlockTypePipelineStepEcho:
			p := primitive.Echo{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		default:
			logger.Error("Unknown step type", "type", stepDefn.GetType())
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		if primitiveError != nil {
			logger.Error("primitive failed", "error", primitiveError)
			if output == nil {
				output = &types.StepOutput{}
			}
			if output.Errors == nil {
				output.Errors = &types.StepErrors{}
			}
			output.Errors.Add(types.StepError{
				Message: primitiveError.Error(),
			})
		}

		// Decorate the errors
		if output.HasErrors() {
			for i := 0; i < len(*output.Errors); i++ {
				err := (*output.Errors)[i]
				err.Step = cmd.StepName
				err.PipelineExecutionID = cmd.PipelineExecutionID
				err.StepExecutionID = cmd.StepExecutionID
				err.Pipeline = pipelineDefn.Name
			}
		}

		// If it's a pipeline step, we need to do something else
		if stepDefn.GetType() == "pipeline" {
			args := types.Input{}
			if cmd.StepInput["args"] != nil {
				args = cmd.StepInput["args"].(map[string]interface{})
			}
			e, err := event.NewPipelineStepStarted(
				event.ForPipelineStepStart(cmd),
				event.WithNewChildPipelineExecutionID(),
				event.WithChildPipeline(cmd.StepInput["name"].(string), args))

			if err != nil {
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return
			}
			err = h.EventBus.Publish(ctx, &e)
			if err != nil {
				logger.Error("Error publishing event", "error", err)
			}
			return
		}

		// All other primitives finish immediately.
		endStep(cmd, output, logger, h, ctx)

	}(ctx, c, h)

	return nil
}

func endStep(cmd *event.PipelineStepStart, output *types.StepOutput, logger *fplog.FlowpipeLogger, h PipelineStepStartHandler, ctx context.Context) {
	e, err := event.NewPipelineStepFinished(
		event.ForPipelineStepStartToPipelineStepFinished(cmd),
		event.WithStepOutput(output))

	if err != nil {
		logger.Error("Error creating Pipeline Step Finished event", "error", err)

		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
		if err2 != nil {
			logger.Error("Error publishing event", "error", err2)
		}
		return
	}

	err = h.EventBus.Publish(ctx, &e)
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}
