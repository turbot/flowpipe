package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus *FpCommandBus
}

type FpCommandBus struct {
	Cb *cqrs.CommandBus
}

// Send sends command to the command bus.
func (c FpCommandBus) Send(ctx context.Context, cmd interface{}) error {

	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, cmd)
	if err != nil {
		return err
	}
	return c.Cb.Send(ctx, cmd)
}

func LogEventMessage(ctx context.Context, cmd interface{}) error {

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

	// Marshal the struct to JSON
	fileData, err := json.Marshal(logMessage) // No indent, single line
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	fileName := path.Join(viper.GetString(constants.ArgLogDir), fmt.Sprintf("%s.jsonl", commandEvent.GetEvent().ExecutionID))

	executionMutex := event.GetEventLogMutex(commandEvent.GetEvent().ExecutionID)
	executionMutex.Lock()
	defer executionMutex.Unlock()

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
