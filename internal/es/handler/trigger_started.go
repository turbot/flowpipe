package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/triggerv2"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
)

type TriggerStarted EventHandler

func (h TriggerStarted) HandlerName() string {
	return execution.TriggerStartedEvent.HandlerName()
}

func (h TriggerStarted) NewEvent() interface{} {
	return &event.TriggerStarted{}
}

func (h TriggerStarted) Handle(ctx context.Context, ei interface{}) error {

	evt, ok := ei.(*event.TriggerStarted)

	if !ok {
		slog.Error("invalid event type", "expected", "*event.TriggerStarted", "actual", ei)
		return perr.BadRequestWithMessage("invalid event type expected *event.TriggerStarted")
	}

	plannerMutex := event.GetEventStoreMutex(evt.Event.ExecutionID)
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	trg, err := db.GetTrigger(evt.Trigger.Name())
	if err != nil {
		slog.Error("Error getting trigger", "error", err)

		plannerMutex.Lock()
		h.raiseError(ctx, evt, err)

		return nil
	}
	triggerRunner := triggerv2.NewTriggerRunner(trg)

	if triggerRunner == nil {
		slog.Error("Error creating trigger runner")

		plannerMutex.Lock()
		h.raiseError(ctx, evt, err)

		return nil
	}

	triggerRunArgs, err := triggerRunner.Validate(evt.Args, nil)
	if err != nil {
		slog.Error("Error validating trigger", "error", err)

		plannerMutex.Lock()
		h.raiseError(ctx, evt, err)

		return nil
	}

	pipelineArgs, err := triggerRunner.GetPipelineArgs(triggerRunArgs)
	if err != nil {
		slog.Error("Error getting pipeline args", "error", err)

		plannerMutex.Lock()
		h.raiseError(ctx, evt, err)

		return nil
	}

	// raise pipeline
	pipelineDefn := trg.Pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewEventForExecutionID(evt.Event.ExecutionID),
		PipelineExecutionID: util.NewPipelineExecutionId(),
		Name:                pipelineName,
		Args:                pipelineArgs,
	}

	slog.Info("Trigger fired", "trigger", trg.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if output.IsServerMode {
		output.RenderServerOutput(ctx, types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, trg.Name(), pipelineName))
	}

	if err := h.CommandBus.Send(ctx, pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)

		if output.IsServerMode {
			output.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error sending pipeline command", err))
		}

		h.raiseError(ctx, evt, err)
	}

	return nil
}

func (h TriggerStarted) raiseError(ctx context.Context, evt *event.TriggerStarted, errToLog error) {
	cmd := event.ExecutionFailFromTriggerStarted(evt, errToLog)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}
}
