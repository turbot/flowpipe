package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepForEachPlanned EventHandler

var stepForEachPlanned = event.StepForEachPlanned{}

func (h StepForEachPlanned) HandlerName() string {
	return stepForEachPlanned.HandlerName()
}

func (StepForEachPlanned) NewEvent() interface{} {
	return &event.StepForEachPlanned{}
}

func (h StepForEachPlanned) Handle(ctx context.Context, ei interface{}) error {
	logger := fplog.Logger(ctx)
	e, ok := ei.(*event.StepForEachPlanned)
	if !ok {
		logger.Error("invalid event type", "expected", "*event.StepForEachPlanned", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.StepForEachPlanned")
	}

	logger.Debug("step_for_each_planned event handler", "event", e)

	for _, nextStep := range e.NextSteps {
		runOneStep(ctx, h.CommandBus, e, &nextStep)
	}
	return nil
}

func runOneStep(ctx context.Context, commandBus *FpCommandBus, e *event.StepForEachPlanned, nextStep *modconfig.NextStep) {

	logger := fplog.Logger(ctx)

	var forEachControl *modconfig.StepForEach

	forEachControl = &modconfig.StepForEach{
		Key: nextStep.StepForEach.Key,
		// Output:     &forEachOutput,
		TotalCount: nextStep.StepForEach.TotalCount,
		Each:       nextStep.StepForEach.Each,
	}

	nextStep.StepForEach = forEachControl

	cmd, err := event.NewPipelineStepQueueFromStepForEachPlanned(e, nextStep)
	if err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}

		return
	}

	if err := commandBus.Send(ctx, cmd); err != nil {
		err := commandBus.Send(ctx, event.NewPipelineFailFromStepForEachPlanned(e, err))
		if err != nil {
			logger.Error("Error publishing event", "error", err)
		}
		return
	}
}
