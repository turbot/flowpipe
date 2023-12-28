package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/pipe-fittings/perr"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus FpCommandBus
}

type FpCommandBus interface {
	Send(ctx context.Context, command interface{}) error
	SendWithLock(ctx context.Context, command interface{}, lock *sync.Mutex) error
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

func (c FpCommandBusImpl) SendWithLock(ctx context.Context, cmd interface{}, lock *sync.Mutex) error {

	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, cmd, lock)
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

	executionID := commandEvent.GetEvent().ExecutionID

	var ex *execution.ExecutionInMemory
	if execution.ExecutionMode == "in-memory" {
		if commandEvent.HandlerName() == "command.pipeline_queue" {
			ex = &execution.ExecutionInMemory{
				ID:                 executionID,
				PipelineExecutions: map[string]*execution.PipelineExecution{},
				Lock:               event.GetEventStoreMutex(executionID),
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

	// Marshal the struct to JSON
	fileData, err := json.Marshal(logMessage) // No indent, single line
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)
		os.Exit(1)
	}

	eventStoreFilePath := filepaths.EventStoreFilePath(executionID)

	if lock == nil {
		executionMutex := event.GetEventStoreMutex(executionID)
		executionMutex.Lock()
		defer executionMutex.Unlock()
	}

	// Append the JSON data to a file
	file, err := os.OpenFile(eventStoreFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return perr.InternalWithMessage("Error opening file " + err.Error())
	}
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return perr.InternalWithMessage("Error writing to file " + err.Error())
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return perr.InternalWithMessage("Error writing to file " + err.Error())
	}

	return nil
}
