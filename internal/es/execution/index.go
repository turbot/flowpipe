package execution

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/types"
)

func LoadEventStoreEntries(executionID string) ([]types.EventLogEntry, error) {

	// Open the JSONL file
	eventStoreFilePath := filepaths.EventStoreFilePath(executionID)
	file, err := os.Open(eventStoreFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize*40), bufio.MaxScanTokenSize*40)

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
			log.Println("Error parsing line:", err)
			continue
		}

		// Append the parsed event to the events slice
		eventLogEntries = append(eventLogEntries, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return eventLogEntries, nil
}
