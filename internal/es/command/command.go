package command

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
)

type CommandHandler struct {
	// Command handlers can only send events, they are not even permitted access
	// to the CommandBus.
	EventBus *FpEventBus
}

type FpEventBus struct {
	Eb *cqrs.EventBus
}

// Publish sends event to the event bus.
func (c FpEventBus) Publish(ctx context.Context, event interface{}) error {
	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, event, nil)
	if err != nil {
		return err
	}

	return c.Eb.Publish(ctx, event)
}

func (c FpEventBus) PublishWithLock(ctx context.Context, event interface{}, lock *sync.Mutex) error {
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

	// Marshal the struct to JSON
	fileData, err := json.Marshal(logMessage) // No indent, single line
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	fileName := path.Join(viper.GetString(constants.ArgLogDir), fmt.Sprintf("%s.jsonl", commandEvent.GetEvent().ExecutionID))

	if lock == nil {
		executionMutex := event.GetEventStoreMutex(commandEvent.GetEvent().ExecutionID)
		executionMutex.Lock()
		defer executionMutex.Unlock()
	}

	// Append the JSON data to a file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
