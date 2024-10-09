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
	plannerMutex.Lock()

	trg, err := db.GetTrigger(evt.Trigger.Name())
	if err != nil {
		slog.Error("Error getting trigger", "error", err)

		fperr := perr.InternalWithMessage("error getting trigger")
		h.raiseError(ctx, evt, fperr)

		return nil
	}
	triggerRunner := triggerv2.NewTriggerRunner(trg, evt.Event.ExecutionID, "")

	if triggerRunner == nil {
		slog.Error("Error creating trigger runner")

		fperr := perr.InternalWithMessage("error creating trigger runner")

		h.raiseError(ctx, evt, fperr)

		return nil
	}

	cmds, err := triggerRunner.ExecuteTriggerWithArgs(ctx, evt.Args, nil)
	if err != nil {
		slog.Error("Error executing trigger", "error", err)

		if output.IsServerMode {
			output.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error executing trigger", err))
		}

		if fperr, ok := err.(perr.ErrorModel); ok {
			h.raiseError(ctx, evt, fperr)
		} else {
			fperr := perr.InternalWithMessage("error executing trigger " + err.Error())
			h.raiseError(ctx, evt, fperr)
		}

		return nil
	}

	if len(cmds) == 0 {
		slog.Info("No commands to execute, ending trigger execution")
		cmd := event.TrigerFinishFromTriggerStarted(evt)
		err := h.CommandBus.Send(ctx, cmd)
		if err != nil {
			slog.Error("Error publishing event", "error", err)
		}

		return nil
	}

	for _, cmd := range cmds {
		if err := h.CommandBus.Send(context.TODO(), cmd); err != nil {
			slog.Error("Error sending pipeline command", "error", err)
			if output.IsServerMode {
				output.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error sending pipeline command", err))
			}

			fperr := perr.InternalWithMessage("error sending pipeline command " + err.Error())
			h.raiseError(ctx, evt, fperr)
		}
	}

	return nil
}

func (h TriggerStarted) raiseError(ctx context.Context, evt *event.TriggerStarted, errToLog perr.ErrorModel) {
	cmd := event.ExecutionFailFromTriggerStarted(evt, errToLog)
	err := h.CommandBus.Send(ctx, cmd)
	if err != nil {
		slog.Error("Error publishing event", "error", err)
	}
}
