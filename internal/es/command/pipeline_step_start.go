package command

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/perr"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

type PipelineStepStartHandler CommandHandler

var pipelineStepStart = event.PipelineStepStart{}

func (h PipelineStepStartHandler) HandlerName() string {
	return pipelineStepStart.HandlerName()
}

func (h PipelineStepStartHandler) NewCommand() interface{} {
	return &event.PipelineStepStart{}
}

// * This is the handler that will actually execute the primitive
// *
// * At the end of the execution it will raise the appropriate event: PipelineStepFinished or PipelineFailed
// *
// * Also note the "special" step handler, this is if it needs to launch a child pipeline
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
		stepOutput := make(map[string]interface{})

		pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

		evalContext, err := ex.BuildEvalContext(pipelineDefn, pe)
		if err != nil {
			logger.Error("Error building eval context while calculating output", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return
		}

		// Check if the step should be skipped. This is determined by the evaluation of the IF clause during the
		// pipeline_plan phase
		if cmd.NextStepAction == modconfig.NextStepActionSkip {
			output := &modconfig.Output{
				Status: "skipped",
			}

			endStep(cmd, output, stepOutput, logger, h, stepDefn, evalContext, ctx)
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
		case schema.BlockTypePipelineStepFunction:
			p := primitive.Function{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepContainer:
			p := primitive.Container{}
			output, primitiveError = p.Run(ctx, cmd.StepInput)
		case schema.BlockTypePipelineStepInput:
			p := primitive.Input{
				ExecutionID:         cmd.Event.ExecutionID,
				PipelineExecutionID: cmd.PipelineExecutionID,
				StepExecutionID:     cmd.StepExecutionID,
			}
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
				Error: perr.InternalWithMessage(primitiveError.Error()),
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

		// We have some special steps that need to be handled differently:
		// Pipeline Step -> launch a new pipeline
		// Input Step -> waiting for external event to resume the pipeline
		shouldReturn := specialStepHandler(ctx, stepDefn, cmd, h, logger)
		if shouldReturn {
			return
		}

		// calculate the output blocks
		// If there's a for_each in the step definition, we need to insert the "each" magic variable
		// so the output can refer to it
		evalContext, stepOutput, shouldReturn = calculateStepConfiguredOutput(ctx, stepDefn, evalContext, cmd, logger, h, err, stepOutput)
		if shouldReturn {
			return
		}

		// All other primitives finish immediately.
		endStep(cmd, output, stepOutput, logger, h, stepDefn, evalContext, ctx)

	}(ctx, c, h)

	return nil
}

// This function mutates stepOutput
func calculateStepConfiguredOutput(ctx context.Context, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext, cmd *event.PipelineStepStart, logger *fplog.FlowpipeLogger, h PipelineStepStartHandler, err error, stepOutput map[string]interface{}) (*hcl.EvalContext, map[string]interface{}, bool) {
	for _, outputConfig := range stepDefn.GetOutputConfig() {
		if outputConfig.UnresolvedValue != nil {

			stepForEach := stepDefn.GetForEach()
			if stepForEach != nil {

				evalContext = execution.AddEachForEach(cmd.StepForEach, evalContext)
			}

			ctyValue, diags := outputConfig.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 && diags.HasErrors() {
				logger.Error("Error calculating output on step start", "error", diags)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return nil, stepOutput, true
			}

			goVal, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				logger.Error("Error converting cty value to Go value for output calculation", "error", err)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return nil, stepOutput, true
			}
			stepOutput[outputConfig.Name] = goVal
		} else {
			stepOutput[outputConfig.Name] = outputConfig.Value
		}
	}
	return evalContext, stepOutput, false
}

// If it's a pipeline step, we need to do something else, we we need to start
// a new pipeline execution for the child pipeline
// If it's an input step, we can't complete the step until the API receives the input's answer
func specialStepHandler(ctx context.Context, stepDefn modconfig.PipelineStep, cmd *event.PipelineStepStart, h PipelineStepStartHandler, logger *fplog.FlowpipeLogger) bool {

	if stepDefn.GetType() == schema.AttributeTypePipeline {
		args := modconfig.Input{}
		if cmd.StepInput[schema.AttributeTypeArgs] != nil {
			args = cmd.StepInput[schema.AttributeTypeArgs].(map[string]interface{})
		}

		e, err := event.NewPipelineStepStarted(
			event.ForPipelineStepStart(cmd),
			event.WithNewChildPipelineExecutionID(),
			event.WithChildPipeline(cmd.StepInput[schema.AttributeTypePipeline].(string), args))

		if cmd.StepForEach != nil {
			e.Key = cmd.StepForEach.Key
		} else {
			e.Key = "0"
		}

		if err != nil {
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
			if err2 != nil {
				logger.Error("Error publishing event", "error", err2)
			}
			return true
		}

		err = h.EventBus.Publish(ctx, e)
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}

		return true
	} else if stepDefn.GetType() == schema.BlockTypeInput {

		logger.Info("input step started, waiting for external response", "step", cmd.StepName, "pipelineExecutionID", cmd.PipelineExecutionID, "executionID", cmd.Event.ExecutionID)
		return true
	}

	return false
}

func endStep(cmd *event.PipelineStepStart, output *modconfig.Output, stepOutput map[string]interface{}, logger *fplog.FlowpipeLogger, h PipelineStepStartHandler, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext, ctx context.Context) {

	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]

	var stepLoop *modconfig.StepLoop
	if loopBlock != nil {
		loopDefn := modconfig.GetLoopDefn(stepDefn.GetType())
		if loopDefn == nil {
			// We should never get here, because the loop block should have been validated
			logger.Error("Unknown loop type", "type", stepDefn.GetType())
		}

		var err error
		evalContext, err = execution.AddStepOutputAsResults(cmd.StepName, output, stepOutput, evalContext)
		if err != nil {
			logger.Error("Error adding step output as results", "error", err)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
			return
		}

		// If this is the first iteration of the loop, the cmd.StepLoop should be nil
		// thus the loop.index in the evaluation context should be 0
		//
		// this allows evaluation such as:
		/*

			step "echo" "echo" {
				text = "foo"

				loop {
					if = loop.index < 2
				}
			}
		**/
		//
		// Because this IF is still part of the Iteration 0 loop.index should be 0
		evalContext := execution.AddLoop(cmd.StepLoop, evalContext)

		// We have to evaluate the loop body here before the index is incremented to determine if the loop should run
		// we will have to re-evulate the loop body again after the index is incremented to get the correct values
		diags := gohcl.DecodeBody(loopBlock, evalContext, loopDefn)
		if len(diags) > 0 {
			logger.Error("Error decoding loop block", "error", diags)
			raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
			return
		}

		// We have to indicate here (before raising the step finish) that this is part of the loop that should be executing, i.e. the step is not actually
		// "finished" yet.
		//
		// Unlike the for_each where we know that there are n number of step executions and the planner launched them all at once, the loop is different.
		//
		// The planner has no idea that the step is not yet finished. We have to tell the planner here that it needs to launch another step execution
		if loopDefn.ShouldRun() {
			// Start with 1 because when we get here the first time, it was the 1st iteration of the loop (index = 0)
			//
			// Unlike the previous evaluation, we are not calculating the input for the NEXT iteration of the loop, so we need to increment the index,
			// do not change the currentIndex to 0
			currentIndex := 1
			if cmd.StepLoop != nil {
				previousIndex := cmd.StepLoop.Index
				currentIndex = previousIndex + 1
			}

			stepLoop = &modconfig.StepLoop{
				Index: currentIndex,
				// Input: &newInput,
			}

			// first we need to evaluate the input for the step, this is to support:
			/*
				step "echo" "echo" {
					text = "iteration: ${loop.index}"
				}
			**/
			// for each of the loop iteration the index changes, so we have to re do it again
			// ensure that we also have the "each" variable here
			evalContext = execution.AddLoop(stepLoop, evalContext)
			evalContext = execution.AddEachForEach(cmd.StepForEach, evalContext)

			reevaluatedInput, err := stepDefn.GetInputs(evalContext)
			if err != nil {
				logger.Error("Error re-evaluating inputs for step", "error", err)
				raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
				return
			}

			// get the new input
			newInput, err := loopDefn.UpdateInput(reevaluatedInput)
			if err != nil {
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, err)))
				if err2 != nil {
					logger.Error("Error publishing event", "error", err2)
				}
				return
			}

			stepLoop.Input = &newInput
		} else {
			stepLoop = cmd.StepLoop
			// complete the loop
			// input is not required here
			stepLoop.Input = nil
			stepLoop.LoopCompleted = true
		}

	}

	e, err := event.NewPipelineStepFinished(
		event.ForPipelineStepStartToPipelineStepFinished(cmd),
		event.WithStepOutput(output, stepOutput, stepLoop))

	if err != nil {
		logger.Error("Error creating Pipeline Step Finished event", "error", err)
		raisePipelineFailedEventFromPipelineStepStart(ctx, h, cmd, err, logger)
		return
	}

	err = h.EventBus.Publish(ctx, e)
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}

func raisePipelineFailedEventFromPipelineStepStart(ctx context.Context, h PipelineStepStartHandler, cmd *event.PipelineStepStart, originalError error, logger *fplog.FlowpipeLogger) {
	err := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepStartToPipelineFailed(cmd, originalError)))
	if err != nil {
		logger.Error("Error publishing event", "error", err)
	}
}
