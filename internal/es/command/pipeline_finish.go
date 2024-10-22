package command

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return execution.PipelineFinishCommand.HandlerName()
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineFinish)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineFinish", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineFinish")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	executionID := cmd.Event.ExecutionID

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(executionID, cmd.PipelineExecutionID)
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}
	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	var output map[string]interface{}
	var outputCalculationErrors []perr.ErrorModel

	if len(pipelineDefn.OutputConfig) > 0 {
		outputBlock := map[string]interface{}{}

		// If all dependencies met, we then calculate the value of this output
		evalContext, err := ex.BuildEvalContext(pipelineDefn, pex)
		if err != nil {
			slog.Error("Error building eval context while calculating output in pipeline_finish", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
		}

		evalContext, err = ex.AddCredentialsToEvalContextFromPipeline(evalContext, pipelineDefn)
		if err != nil {
			slog.Error("Error adding credentials to eval context while calculating output in pipeline_finish", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
		}

		evalContext, err = ex.AddConnectionsToEvalContextFromPipeline(evalContext, pipelineDefn)
		if err != nil {
			slog.Error("Error adding connections to eval context while calculating output in pipeline_finish", "error", err)
			err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
			if err2 != nil {
				slog.Error("Error publishing PipelineFailed event", "error", err2)
			}
		}

		for _, output := range pipelineDefn.OutputConfig {
			ctyValue, diags := output.UnresolvedValue.Value(evalContext)
			if len(diags) > 0 {
				err := error_helpers.HclDiagsToError("output", diags)
				slog.Error("Error calculating output "+output.Name, "error", err)

				outputCalculationErrors = append(outputCalculationErrors, perr.InternalWithMessage("Error calculating output '"+output.Name+"': "+err.Error()))
				continue
			}
			val, err := hclhelpers.CtyToGo(ctyValue)
			if err != nil {
				slog.Error("Error converting cty value to Go value", "error", err)
				return err
			}
			outputBlock[output.Name] = val
		}
		output = outputBlock
	}

	if len(outputCalculationErrors) > 0 {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx,
			event.PipelineFailedWithEvent(event.NewFlowEvent(cmd.Event)),
			event.PipelineFailedWithMultipleErrors(cmd.PipelineExecutionID, pipelineDefn.Name(), outputCalculationErrors),
			event.PipelineFailedWithOutput(output)))

		if err2 != nil {
			slog.Error("Error publishing pipeline_failed event", "error", err2)
		}
		return nil
	}

	e, err := event.NewPipelineFinished(event.ForPipelineFinish(cmd, output))
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("Error publishing pipeline_failed event", "error", err2)
		}
		return nil
	}

	err = h.EventBus.Publish(ctx, e)
	if err != nil {
		slog.Error("Error publishing pipeline_finished event", "error", err)
	}
	return nil
}
