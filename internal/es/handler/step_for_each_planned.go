package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"log/slog"
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
	e, ok := ei.(*event.StepForEachPlanned)
	if !ok {
		slog.Error("invalid event type", "expected", "*event.StepForEachPlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepForEachPlanned")
	}

	slog.Debug("step_for_each_planned event handler", "event", e)

	if len(e.NextSteps) == 0 {
		slog.Debug("step_for_each_planned event handler - no next steps")

		// If nothing is planned, then we're done for this "step_for_each" step. Run the pipeline planner (not the step_for_each_planner)
		cmd := event.NewPipelinePlanFromStepForEachPlanned(e)
		return h.CommandBus.Send(ctx, cmd)
	}

	for i := range e.NextSteps {
		nextStep := e.NextSteps[i]
		runOneStep(ctx, h.CommandBus, e, &nextStep)
	}
	return nil
}

func runOneStep(ctx context.Context, commandBus *FpCommandBus, e *event.StepForEachPlanned, nextStep *modconfig.NextStep) {

		// forEachControl := &modconfig.StepForEach{
		// 	ForEachStep: true,
		// 	Key:         nextStep.StepForEach.Key,
		// 	// Output:     &forEachOutput,
		// 	TotalCount: nextStep.StepForEach.TotalCount,
		// 	Each:       nextStep.StepForEach.Each,
		// }

		// nextStep.StepForEach = forEachControl

		cmd, err := event.NewStepQueueFromStepForEachPlanned(e, nextStep)
	if err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return
	}

	if err := commandBus.Send(ctx, cmd); err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}
		return
	}
}
