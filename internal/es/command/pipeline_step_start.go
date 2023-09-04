package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
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

		// Check if the step should be skipped. This is determined by the evaluation of the IF clause during the
		// pipeline_plan phase
		if cmd.NextStepAction == modconfig.NextStepActionSkip {
			output := &modconfig.Output{
				Status: "skipped",
			}

			endStep(cmd, output, logger, h, ctx)
			return
		}

		var output *modconfig.Output
		var primitiveError error
		switch stepDefn.GetType() {
		case schema.BlockTypePipelineStepExec:
			p := primitive.Exec{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepHttp:
			p := primitive.HTTPRequest{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepPipeline:
			p := primitive.RunPipeline{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepEmail:
			p := primitive.Email{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepQuery:
			p := primitive.Query{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepSleep:
			p := primitive.Sleep{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
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
				output = &modconfig.Output{}
			}
			if output.Errors == nil {
				output.Errors = []modconfig.StepError{}
			}

			output.Errors = append(output.Errors, modconfig.StepError{
				Message: primitiveError.Error(),
			})

		}

		// Decorate the errors
		if output.HasErrors() {
			output.Status = "failed"
			for i := 0; i < len(output.Errors); i++ {
				(output.Errors)[i].Step = cmd.StepName
				(output.Errors)[i].PipelineExecutionID = cmd.PipelineExecutionID
				(output.Errors)[i].StepExecutionID = cmd.StepExecutionID
				(output.Errors)[i].Pipeline = pipelineDefn.Name()
			}
		} else {
			output.Status = "finished"
		}

		// If it's a pipeline step, we need to do something else, we we need to start
		// a new pipeline execution for the child pipeline
		if stepDefn.GetType() == schema.AttributeTypePipeline {
			args := modconfig.Input{}
			if cmd.StepInput[schema.AttributeTypeArgs] != nil {
				args = cmd.StepInput[schema.AttributeTypeArgs].(map[string]interface{})
			}

			e, err := event.NewPipelineStepStarted(
				event.ForPipelineStepStart(cmd),
				event.WithNewChildPipelineExecutionID(),
				event.WithChildPipeline(cmd.StepInput[schema.AttributeTypePipeline].(string), args))

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

func endStep(cmd *event.PipelineStepStart, output *modconfig.Output, logger *fplog.FlowpipeLogger, h PipelineStepStartHandler, ctx context.Context) {
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

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		logger.Error("error1", "error", err)
	}

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		logger.Error("Error getting pipeline definiton", "error", err)
	}

	stepExecution := ex.PipelineExecutions[cmd.PipelineExecutionID].StepExecutions[cmd.StepExecutionID]

	stepDefn := pipelineDefn.GetStep(stepExecution.Name)
	if err != nil {
		logger.Error("Error getting step definition", "error", err)
	}

	outputBlock := map[string]interface{}{}
	for k, v := range stepDefn.GetOutputConfig() {
		outputBlock[k] = v.Value
	}

	e.StepOutput = outputBlock

	err = h.EventBus.Publish(ctx, &e)
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}
