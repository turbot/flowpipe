package execution

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"strconv"

	"github.com/turbot/flowpipe/internal/filepaths"

	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
)

// TODO: make this better
var Mode string

func ServerOutput(output string) {
	if Mode == "server" {
		fmt.Printf("%s %s\n", time.Now().Format(time.RFC3339), output) //nolint:forbidigo // Output
	}
}

func LoadEventStoreEntries(executionID string) ([]types.EventLogEntry, error) {

	// Open the JSONL file
	eventStoreFilePath := filepaths.EventStoreFilePath(executionID)
	file, err := os.Open(eventStoreFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, constants.MaxScanSize), constants.MaxScanSize)

	// Create a slice to hold the parsed eventLogEntries
	var eventLogEntries []types.EventLogEntry

	// Read the file line by line
	for scanner.Scan() {
		line := scanner.Bytes()

		// Create a new Event struct to hold the parsed event
		var event types.EventLogEntry

		// Parse the line into the Event struct
		err := json.Unmarshal(line, &event)
		if err != nil {
			slog.Error("Error parsing line:", "error", err)
			continue
		}

		// Append the parsed event to the events slice
		eventLogEntries = append(eventLogEntries, event)
	}

	if err := scanner.Err(); err != nil {
		if err.Error() == bufio.ErrTooLong.Error() {
			return nil, perr.InternalWithMessageAndType(perr.ErrorCodeInternalTokenTooLarge, "Event log entry too large. Max size is "+strconv.Itoa(constants.MaxScanSize))
		}
		return nil, err
	}

	return eventLogEntries, nil
}
