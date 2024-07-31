package command

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/pipe-fittings/perr"
)

type CommandHandler struct {
	// Command handlers can only send events, they are not even permitted access
	// to the CommandBus.
	EventBus FpEventBus
}

type FpEventBus interface {
	Publish(ctx context.Context, event interface{}) error
}

type FpEventBusImpl struct {
	Eb *cqrs.EventBus
}

// Publish sends event to the event bus.
func (c FpEventBusImpl) Publish(ctx context.Context, event interface{}) error {
	// Unfortunately we need to save the event *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, event, nil)
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

	logMessage := event.NewEventLogFromCommand(commandEvent)

	executionID := commandEvent.GetEvent().ExecutionID

	var ex *execution.ExecutionInMemory
	var err error

	ex, err = execution.GetExecution(executionID)
	if err != nil {
		slog.Error("Error getting execution", "error", err)
		return perr.InternalWithMessage("Error getting execution")
	}

	err = ex.AddEvent(logMessage)
	if err != nil {
		slog.Error("Error adding event to execution", "error", err)
		return perr.InternalWithMessage("Error adding event to execution")
	}

	if strings.ToLower(os.Getenv("FLOWPIPE_EVENT_FORMAT")) == "jsonl" {
		err := execution.LogEventMessageToFile(ctx, logMessage)
		if err != nil {
			return err
		}
	}

	db, err := store.OpenFlowpipeDB()
	if err != nil {
		return perr.InternalWithMessage("Error opening SQLite database " + err.Error())
	}
	defer db.Close()

	err = execution.SaveEventToSQLite(db, executionID, logMessage)
	if err != nil {
		slog.Error("Error saving event to SQLite", "error", err)
		return err
	}
	return nil
}
