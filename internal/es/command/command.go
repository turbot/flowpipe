package command

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type CommandHandler struct {
	// Command handlers can only send events, they are not even permitted access
	// to the CommandBus.
	EventBus FpEventBus
}

type FpEventBus interface {
	Publish(ctx context.Context, event interface{}) error
	PublishWithLock(ctx context.Context, event interface{}, lock *sync.Mutex) error
}

type FpEventBusImpl struct {
	Eb *cqrs.EventBus
}

// Publish sends event to the event bus.
func (c FpEventBusImpl) Publish(ctx context.Context, event interface{}) error {
	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, event, nil)
	if err != nil {
		return err
	}

	return c.Eb.Publish(ctx, event)
}

func (c FpEventBusImpl) PublishWithLock(ctx context.Context, event interface{}, lock *sync.Mutex) error {
	// Unfortunately we need to save the event log *before* we send this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, event, lock)
	if err != nil {
		return err
	}

	return c.Eb.Publish(ctx, event)
}

func LogEventMessage(ctx context.Context, evt interface{}, lock *sync.Mutex) error {

	commandEvent, ok := evt.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	logMessage := event.EventLogEntry{
		Level:     "info",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Caller:    "command",
		Message:   "es",
		EventType: commandEvent.HandlerName(),
		Payload:   evt,
	}

	executionID := commandEvent.GetEvent().ExecutionID

	var ex *execution.ExecutionInMemory
	var err error

	ex, err = execution.GetExecution(executionID)
	if err != nil {
		slog.Error("Error getting execution", "error", err)
		return perr.InternalWithMessage("Error getting execution")
	}

	if lock == nil {
		ex.Lock.Lock()
		defer ex.Lock.Unlock()
	}

	err = ex.AddEvent(logMessage)
	if err != nil {
		slog.Error("Error adding event to execution", "error", err)
		return perr.InternalWithMessage("Error adding event to execution")
	}
	return nil
}
