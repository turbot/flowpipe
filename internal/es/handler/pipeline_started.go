package handler

import (
	"context"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
	"log/slog"
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

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutput(e.Event.CreatedAt, "pipeline", "started"),
			e.Event.ExecutionID, "")
		output.RenderServerOutput(ctx, p)
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStarted(e))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStartedToPipelineFail(e, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
