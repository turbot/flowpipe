package handler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus FpCommandBus
}

type FpCommandBus interface {
	Send(ctx context.Context, command interface{}) error
}

type FpCommandBusImpl struct {
	Cb *cqrs.CommandBus
}

// Send sends command to the command bus.
func (c FpCommandBusImpl) Send(ctx context.Context, cmd interface{}) error {

	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, cmd, nil)
	if err != nil {
		return err
	}
	return c.Cb.Send(ctx, cmd)
}

func LogEventMessage(ctx context.Context, cmd interface{}, lock *sync.Mutex) error {

	commandEvent, ok := cmd.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	logMessage := event.EventLogEntry{
		Level:     "info",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Caller:    "command",
		Message:   "es",
		EventType: commandEvent.HandlerName(),
		Payload:   commandEvent,
	}

	newExecution := false

	pipelineQueueCmd, ok := commandEvent.(*event.PipelineQueue)
	if ok && pipelineQueueCmd.ParentStepExecutionID == "" {
		newExecution = true
	}

	var ex *execution.ExecutionInMemory
	executionID := commandEvent.GetEvent().ExecutionID
	if newExecution {
		ex = &execution.ExecutionInMemory{
			Execution: execution.Execution{
				ID:                 executionID,
				PipelineExecutions: map[string]*execution.PipelineExecution{},
				Lock:               event.GetEventStoreMutex(executionID),
			},
		}

		// Effectively forever
		ok := cache.GetCache().SetWithTTL(executionID, ex, 10*365*24*time.Hour)
		if !ok {
			slog.Error("Error setting execution in cache", "execution_id", executionID)
			return perr.InternalWithMessage("Error setting execution in cache")
		}
	} else {
		var err error
		ex, err = execution.GetExecution(executionID)
		if err != nil {
			slog.Error("Error getting execution from cache", "execution_id", executionID)
			return perr.InternalWithMessage("Error getting execution from cache")
		}
	}

	err := ex.AddEvent(logMessage)
	if err != nil {
		slog.Error("Error adding event to execution", "error", err)
		return err
	}
	return nil
}
