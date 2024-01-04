package event

import (
	"sync"

	"github.com/turbot/flowpipe/internal/metrics"
)

// Define a struct that represents your JSON structure
type EventLogEntry struct {
	Level     string      `json:"level"`
	Timestamp string      `json:"ts"`
	Caller    string      `json:"caller"`
	Message   string      `json:"msg"`
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
}

var eventStoreWriteMutexes = make(map[string]*sync.Mutex)
var eventStoreWriteMutexLock sync.Mutex

func GetEventStoreMutex(executionId string) *sync.Mutex {
	eventStoreWriteMutexLock.Lock()
	defer eventStoreWriteMutexLock.Unlock()

	if mutex, exists := eventStoreWriteMutexes[executionId]; exists {
		return mutex
	} else {
		newMutex := &sync.Mutex{}
		eventStoreWriteMutexes[executionId] = newMutex
		return newMutex
	}
}

func ReleaseEventLogMutex(executionId string) {
	eventStoreWriteMutexLock.Lock()
	defer eventStoreWriteMutexLock.Unlock()

	delete(eventStoreWriteMutexes, executionId)
	metrics.RunMetricInstance.EndExecution(executionId)
}
