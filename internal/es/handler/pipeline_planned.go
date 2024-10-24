package handler

import (
	"context"
	"os"
	"time"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type PipelinePlanned EventHandler

func (h PipelinePlanned) HandlerName() string {
	return execution.PipelinePlannedEvent.HandlerName()
}

func (PipelinePlanned) NewEvent() interface{} {
	return &event.PipelinePlanned{}
}

func (h PipelinePlanned) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.PipelinePlanned)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelinePlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelinePlanned")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
	if err != nil {
		slog.Error("pipeline_planned: Error loading pipeline execution", "error", err)
		err2 := h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(evt, err))
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

	if len(evt.NextSteps) == 0 {
		// PRE: No new steps to execute, so the planner should just check to see if
		// all existing steps are complete.
		if pex.IsComplete() {
			if pex.ShouldFail() {
				// There's no error supplied here because it's the step failure that is causing the pipeline to fail
				cmd := event.NewPipelineFailFromPipelinePlanned(evt, nil)
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(evt, err))
				}
				return h.CommandBus.Send(ctx, cmd)
			} else {
				cmd, err := event.NewPipelineFinish(event.ForPipelinePlannedToPipelineFinish(evt))
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(evt, err))
				}
				return h.CommandBus.Send(ctx, cmd)
			}
		} else {
			// There are no new steps to run, but the pipeline isn't complete, so we need to wait
			// for the existing steps to complete.
			//
			// We need to figure out if the steps are still running or all of the steps are in "steady state" waiting
			// for external events to trigger them. Currently, this only applies to input steps
			//
			// And also it applies to pipeline step where the pipeline is currently in paused state
			onlyInputStepsRunning := false
			var latestActionTimestamp time.Time
			for _, stepExecution := range pex.StepExecutions {
				slog.Debug("Checking step", "step", stepExecution.Name, "status", stepExecution.Status)
				if stepExecution.Status == "starting" || stepExecution.Status == "started" {
					stepName := stepExecution.Name
					stepDefn := pipelineDefn.GetStep(stepName)
					if stepDefn.GetType() == schema.BlockTypePipelineStepInput {
						onlyInputStepsRunning = true
						if latestActionTimestamp.IsZero() || stepExecution.StartTime.After(latestActionTimestamp) {
							latestActionTimestamp = stepExecution.StartTime
						}
					} else if stepDefn.GetType() == schema.BlockTypePipelineStepPipeline {
						slog.Debug("pipeline step, checking status of pipeline")
						childPex := ex.FindPipelineExecutionByItsParentStepExecution(stepExecution.ID)
						if childPex == nil {
							slog.Error("pipeline step has no child pipeline execution ... may not started yet", "step", stepName)
							onlyInputStepsRunning = false
							break
						}

						if childPex.IsPaused() {
							slog.Info("child pipeline is paused", "pipeline", pipelineDefn.Name(), "child_pipeline", childPex.Name,
								"status", childPex.Status, "childPipelineExecutionId", childPex.ID, "stepExecutionId", stepExecution.ID)
							onlyInputStepsRunning = true

							if latestActionTimestamp.IsZero() || stepExecution.StartTime.After(latestActionTimestamp) {
								latestActionTimestamp = stepExecution.StartTime
							}

							continue
						} else {
							onlyInputStepsRunning = false
							// As soon as there's another non-input step that is running, we break
							break
						}
					} else {
						onlyInputStepsRunning = false
						// As soon as there's another non-input step that is running, we break
						break
					}
				}
			}

			if onlyInputStepsRunning {

				if !pex.ResumedAt.IsZero() {
					if pex.ResumedAt.After(latestActionTimestamp) {
						latestActionTimestamp = pex.ResumedAt
					}
				}

				slog.Info("Pipeline is waiting for steps to complete", "pipeline", pipelineDefn.Name(), "onlyInputStepsRunning", onlyInputStepsRunning)

				// Auto paused after 1 minute of inaction (waiting for input steps to complete)
				// TODO: make this configurable
				timeoutStr := os.Getenv("FLOWPIPE_AUTO_PAUSE_DURATION")
				if timeoutStr == "" {
					timeoutStr = "60s"
				}
				timeout, err := time.ParseDuration(timeoutStr)
				if err != nil {
					timeout = 60 * time.Second
				}

				if time.Since(latestActionTimestamp) > timeout {
					slog.Info("Pipeline has been waiting for input steps to complete for more than 1 minutes. Automatically pausing the pipeline.", "pipeline", pipelineDefn.Name(), "onlyInputStepsRunning", onlyInputStepsRunning)
					cmd := event.PipelinePauseFromPipelinePlanned(evt)
					err := h.CommandBus.Send(ctx, cmd)
					if err != nil {
						slog.Error("Error publishing event", "error", err)
					}
				} else {
					go func() {
						// Check every 5 second what the status of the pipeline is
						time.Sleep(5 * time.Second)
						slog.Info("Pipeline has been waiting for input steps to complete for more than 5 seconds, raising pipeline plan event", "pipeline", pipelineDefn.Name(), "onlyInputStepsRunning", onlyInputStepsRunning)
						cmd := event.PipelinePlanFromPipelinePlanned(evt)
						if err != nil {
							slog.Error("Error publishing event", "error", err)
							return
						}

						err := h.CommandBus.Send(ctx, cmd)
						if err != nil {
							slog.Error("Error publishing event", "error", err)
						}
					}()
				}
			}

		}

		return nil
	}

	// Check if there is a step that is "inaccessible", if so then we terminate the pipeline
	// since there's no possibility of it ever completing
	// TODO: there's optimisation here, we could potentially run all the other steps that can run
	// TODO: but for now take the simplest route
	pipelineInaccessible := false
	for _, nextStep := range evt.NextSteps {
		if nextStep.Action == modconfig.NextStepActionInaccessible {
			pipelineInaccessible = true
			break
		}
	}

	if pipelineInaccessible {
		slog.Info("Pipeline is inaccessible, terminating", "pipeline", pipelineDefn.Name())
		// TODO: what is the error on the pipeline?
		cmd := event.NewPipelineFailFromPipelinePlanned(evt, nil)
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(evt, err))
		}
		return h.CommandBus.Send(ctx, cmd)
	}

	// The planner has told us what steps to run next, our job is to start them
	for _, nextStep := range evt.NextSteps {

		stepDefn := pipelineDefn.GetStep(nextStep.StepName)
		stepForEach := stepDefn.GetForEach()

		// If there's a for each, runs a new command: step_for_each_planner
		if !helpers.IsNil(stepForEach) {
			var err error

			stepForEachPlanCmd := event.NewStepForEachPlanFromPipelinePlanned(evt, nextStep.StepName)
			err = h.CommandBus.Send(ctx, stepForEachPlanCmd)

			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(evt, err))
			}

			// don't return here, but process the next step
			continue
		}

		var stepLoop *modconfig.StepLoop
		if !helpers.IsNil(stepDefn.GetLoopConfig()) {
			stepLoop = &modconfig.StepLoop{
				Index: 0,
			}
		}

		// Start each step in parallel
		runNonForEachStep(ctx, h.CommandBus, evt, modconfig.Output{}, nextStep.Action, nextStep, nextStep.Input, stepLoop)
	}

	return nil
}

func runNonForEachStep(ctx context.Context, commandBus FpCommandBus, e *event.PipelinePlanned, forEachOutput modconfig.Output, forEachNextStepAction modconfig.NextStepAction, nextStep modconfig.NextStep, input modconfig.Input, stepLoop *modconfig.StepLoop) {

	// If a step does not have a for_each, we still build a for_each control but with key of "0"
	forEachControl := &modconfig.StepForEach{
		ForEachStep: false,
		Key:         "0",
		TotalCount:  1,
		Each:        json.SimpleJSONValue{Value: cty.StringVal("0")},
	}

	cmd, err := event.NewStepQueue(event.StepQueueForPipelinePlanned(e), event.StepQueueWithStep(nextStep.StepName, input, forEachControl, nextStep.StepLoop, forEachNextStepAction))
	cmd.StepLoop = stepLoop

	if err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return
	}

	if err := commandBus.Send(ctx, cmd); err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromPipelinePlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return
	}
}
