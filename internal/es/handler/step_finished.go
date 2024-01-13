package handler

import (
	"context"
	"github.com/turbot/pipe-fittings/utils"
	"log/slog"
	"time"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type StepFinished EventHandler

func (h StepFinished) HandlerName() string {
	return execution.StepFinishedEvent.HandlerName()
}

func (StepFinished) NewEvent() interface{} {
	return &event.StepFinished{}
}

// This is the generic step finish event handler that is fired by the step_start command
//
// Do not confuse this with pipeline_step_finish **command** which is raised when a child pipeline has finished
func (h StepFinished) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.StepFinished)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepFinished", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepFinished")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("step_finished: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	pex := ex.PipelineExecutions[evt.PipelineExecutionID]

	// If the pipeline has been canceled or paused, then no planning is required as no
	// more work should be done.
	if pex.IsCanceled() || pex.IsPaused() || pex.IsFinishing() || pex.IsFinished() {
		return nil
	}

	stepExecution := pex.StepExecutions[evt.StepExecutionID]
	stepName := stepExecution.Name

	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
	}

	stepDefn := pipelineDefn.GetStep(stepName)
	if stepDefn == nil {
		slog.Error("step not found", "step_name", stepName)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, perr.BadRequestWithMessage("step not found"))))
	}

	// Check if we are in a retry block
	if evt.StepRetry != nil && !evt.StepRetry.RetryCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForRetry(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	} else if evt.StepRetry != nil && evt.StepRetry.RetryCompleted {
		// this means we have an error BUT the retry has been exhausted, run the planner
		if output.IsServerMode {
			feKey, li, ri := getIndices(evt)
			sp := types.NewServerOutputPrefixWithExecId(evt.Event.CreatedAt, "pipeline", &evt.Event.ExecutionID)
			prefix := types.NewParsedEventPrefix(pipelineDefn.PipelineName, &stepName, feKey, li, ri, &sp)
			pe := types.NewParsedEvent(prefix, evt.Event.ExecutionID, h.HandlerName(), stepDefn.GetType(), "")
			duration := utils.HumanizeDuration(time.Now().Sub(stepExecution.StartTime))
			output.RenderServerOutput(ctx, types.NewParsedErrorEvent(pe, evt.Output.Errors, evt.Output.Data, &duration, false, true))
		}
		cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(evt))
		if err != nil {
			slog.Error("error creating pipeline_plan command", "error", err)
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// First thing first .. before we run the planner (either pipeline plan or step for each plan),
	// check if we are in a loop. If we are in a loop start the next loop
	loopBlock := stepDefn.GetUnresolvedBodies()[schema.BlockTypeLoop]
	if loopBlock != nil && evt.StepLoop != nil && !evt.StepLoop.LoopCompleted {
		cmd := event.NewStepQueueFromPipelineStepFinishedForLoop(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	// If the step is a for each step, run the for each planner, not the pipeline planner
	if !helpers.IsNil(stepDefn.GetForEach()) {
		cmd := event.NewStepForEachPlanFromPipelineStepFinished(evt, stepName)
		return h.CommandBus.Send(ctx, cmd)
	}

	if output.IsServerMode && evt.Output != nil && evt.Output.Status != "skipped" {
		feKey, li, ri := getIndices(evt)
		sp := types.NewServerOutputPrefixWithExecId(evt.Event.CreatedAt, "pipeline", &evt.Event.ExecutionID)
		prefix := types.NewParsedEventPrefix(pipelineDefn.PipelineName, &stepName, feKey, li, ri, &sp)
		pe := types.NewParsedEvent(prefix, evt.Event.ExecutionID, h.HandlerName(), stepDefn.GetType(), "")
		duration := utils.HumanizeDuration(time.Now().Sub(stepExecution.StartTime))
		switch evt.Output.Status {
		case "finished":
			output.RenderServerOutput(ctx, types.NewParsedEventWithOutput(pe, evt.Output.Data, evt.StepOutput, &duration, false))
		case "failed":
			rc := true
			if evt.StepRetry != nil {
				rc = evt.StepRetry.RetryCompleted
			}
			output.RenderServerOutput(ctx, types.NewParsedErrorEvent(pe, evt.Output.Errors, evt.Output.Data, &duration, false, rc))
		}
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStepFinished(evt))
	if err != nil {
		slog.Error("error creating pipeline_plan command", "error", err)
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStepFinishedToPipelineFail(evt, err)))
	}

	return h.CommandBus.Send(ctx, cmd)
}

func getIndices(evt *event.StepFinished) (*string, *int, *int) {
	var feKey *string
	var li, ri *int
	if evt.StepForEach != nil && evt.StepForEach.ForEachStep {
		feKey = &evt.StepForEach.Key
	}
	if evt.StepLoop != nil {
		i := evt.StepLoop.Index - 1
		li = &i
	}
	if evt.StepRetry != nil {
		ri = &evt.StepRetry.Count
	}

	return feKey, li, ri
}
