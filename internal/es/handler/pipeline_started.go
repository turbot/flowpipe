package handler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
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
	evt, ok := ei.(*event.PipelineStarted)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.PipelineStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.PipelineStarted")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	if output.IsServerMode {
		pipelineName := ""
		ex, pipelineDefn, err := execution.GetPipelineDefnFromExecution(evt.Event.ExecutionID, evt.PipelineExecutionID)
		if err != nil {
			slog.Error("pipeline_started: error loading pipeline definition from execution", "error", err)
		} else {
			pipelineName = pipelineDefn.PipelineName
		}

		var args map[string]any
		pex := ex.PipelineExecutions[evt.PipelineExecutionID]
		if pex != nil {
			args = pex.Args
		}
		sp := types.NewServerOutputPrefixWithExecId(evt.Event.CreatedAt, "pipeline", &evt.Event.ExecutionID)
		prefix := types.NewPrefixWithServer(pipelineName, sp)
		pe := types.NewParsedEvent(prefix, evt.Event.ExecutionID, event.HandlerPipelineStarted, "", "")
		o := types.NewParsedEventWithInput(pe, args, false)
		output.RenderServerOutput(ctx, o)
	}

	cmd, err := event.NewPipelinePlan(event.ForPipelineStarted(evt))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineStartedToPipelineFail(evt, err)))
	}
	return h.CommandBus.Send(ctx, cmd)
}
