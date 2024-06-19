package execution

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/pipe-fittings/perr"
)

func LogEventMessageToFile(ctx context.Context, logEntry *event.EventLogEntry) error {

	commandEvent, ok := logEntry.Payload.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	fileData, err := json.Marshal(logEntry)
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)
		os.Exit(1)
	}

	eventStoreFilePath := filepaths.EventStoreFilePath(commandEvent.GetEvent().ExecutionID)

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
