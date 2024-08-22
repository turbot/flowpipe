package event

import (
	"sync"
	"time"

	"github.com/turbot/flowpipe/internal/metrics"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/utils"
)

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

type EventLogImpl struct {
	StructVersion string      `json:"struct_version"`
	ID            string      `json:"id"`
	ProcessID     string      `json:"process_id"`
	Message       string      `json:"message"`
	Level         string      `json:"level"`
	CreatedAt     time.Time   `json:"created_at"`
	Detail        interface{} `json:"detail"`
}

func (e *EventLogImpl) GetID() string {
	return e.ID
}

func (e *EventLogImpl) GetStructVersion() string {
	return "2.0"
}

func (e *EventLogImpl) GetEventType() string {
	return e.Message
}

func (e *EventLogImpl) GetDetail() interface{} {
	return e.Detail
}

func (e *EventLogImpl) GetCreatedAt() time.Time {
	return e.CreatedAt.UTC()
}

func (e *EventLogImpl) GetLevel() string {
	return e.Level
}

func (e *EventLogImpl) GetCreatedAtString() string {
	return e.CreatedAt.UTC().Format(utils.RFC3339WithMS)
}

func (e *EventLogImpl) SetCreatedAtString(createdAt string) error {
	ts, err := time.Parse(utils.RFC3339WithMS, createdAt)
	if err != nil {
		return err
	}

	e.CreatedAt = ts
	return nil
}

func (e *EventLogImpl) SetDetail(detail interface{}) {
	e.Detail = detail
}

func NewEventLog() EventLogImpl {
	et := EventLogImpl{
		StructVersion: "2.0",
		Level:         "event",
	}

	return et
}

func NewEventLogFromCommand(command CommandEvent) EventLogImpl {
	et := EventLogImpl{
		StructVersion: "2.0",
		ID:            util.NewProcessLogId(),
		ProcessID:     command.GetEvent().ExecutionID,
		Message:       command.HandlerName(),
		Level:         "event",
		CreatedAt:     command.GetEvent().CreatedAt,
		Detail:        command,
	}

	return et
}
