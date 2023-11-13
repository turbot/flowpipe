package event

import "sync"

// Define a struct that represents your JSON structure
type EventLogEntry struct {
	Level     string      `json:"level"`
	Timestamp string      `json:"ts"`
	Caller    string      `json:"caller"`
	Message   string      `json:"msg"`
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
}

var eventLogWriteMutexes = make(map[string]*sync.Mutex)
var eventLogWriteMutexLock sync.Mutex

func GetMutex(executionId string) *sync.Mutex {
	eventLogWriteMutexLock.Lock()
	defer eventLogWriteMutexLock.Unlock()

	if mutex, exists := eventLogWriteMutexes[executionId]; exists {
		return mutex
	} else {
		newMutex := &sync.Mutex{}
		eventLogWriteMutexes[executionId] = newMutex
		return newMutex
	}
}

func ReleaseMutex(executionId string) {
	eventLogWriteMutexLock.Lock()
	defer eventLogWriteMutexLock.Unlock()

	delete(eventLogWriteMutexes, executionId)
}
