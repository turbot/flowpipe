package handler

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
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

	logger.Debug("pipeline_canceled event handler", "event", e)
	return nil
}
