package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineQueued EventHandler

func (h PipelineQueued) HandlerName() string {
	return execution.PipelineQueuedEvent.HandlerName()
}

func (PipelineQueued) NewEvent() interface{} {
	return &event.PipelineQueued{}
}

// Path from here:
// * PipelineQueued -> PipelineLoad command -> PipelineLoaded event handler
func (h PipelineQueued) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.PipelineQueued)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineQueued", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineQueued")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	if output.IsServerMode {
		p := types.NewServerOutputPipelineExecution(
			types.NewServerOutputPrefix(evt.Event.CreatedAt, "pipeline"),
			evt.Event.ExecutionID, evt.Name, "queued")
		output.RenderServerOutput(ctx, p)
	}

	cmd, err := event.NewPipelineLoad(event.ForPipelineQueued(evt))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineQueuedToPipelineFail(evt, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
