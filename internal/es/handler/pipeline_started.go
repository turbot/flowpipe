package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStarted EventHandler

func (h PipelineStarted) HandlerName() string {
	return execution.PipelineStartedEvent.HandlerName()
}

func (PipelineStarted) NewEvent() interface{} {
	return &event.PipelineStarted{}
}

func (h PipelineStarted) Handle(ctx context.Context, ei interface{}) error {
	e, ok := ei.(*event.PipelineStarted)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineStarted")
	}

	execution.ServerOutput(fmt.Sprintf("[%s] Pipeline started", e.Event.ExecutionID))

	cmd, err := event.NewPipelinePlan(event.ForPipelineStarted(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStartedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
