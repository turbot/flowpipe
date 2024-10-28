package handler

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepForEachPlanned EventHandler

func (h StepForEachPlanned) HandlerName() string {
	return execution.StepForEachPlannedEvent.HandlerName()
}

func (StepForEachPlanned) NewEvent() interface{} {
	return &event.StepForEachPlanned{}
}

func (h StepForEachPlanned) Handle(ctx context.Context, ei interface{}) error {
	evt, ok := ei.(*event.StepForEachPlanned)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepForEachPlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepForEachPlanned")
	}

	slog.Debug("step_for_each_planned event handler", "event", evt)

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	if len(evt.NextSteps) == 0 {
		slog.Debug("step_for_each_planned event handler - no next steps")

		// If nothing is planned, then we're done for this "step_for_each" step. Run the pipeline planner (not the step_for_each_planner)
		cmd := event.NewPipelinePlanFromStepForEachPlanned(evt)
		return h.CommandBus.Send(ctx, cmd)
	}

	for i := range evt.NextSteps {
		nextStep := evt.NextSteps[i]
		runOneStep(ctx, h.CommandBus, evt, &nextStep)
	}
	return nil
}

func runOneStep(ctx context.Context, commandBus FpCommandBus, e *event.StepForEachPlanned, nextStep *modconfig.NextStep) {

	cmd, err := event.NewStepQueueFromStepForEachPlanned(e, nextStep)
	if err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return
	}

	cmd.MaxConcurrency = nextStep.MaxConcurrency

	if err := commandBus.Send(ctx, cmd); err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return
	}
}
