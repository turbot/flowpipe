package event

type ExecutionQueued struct {
	// Event metadata
	Event *Event `json:"event"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"` // PipelineQueue not supported yet

	TriggerQueue *TriggerQueue `json:"trigger_queue"`
}

func (e *ExecutionQueued) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionQueued) HandlerName() string {
	return HandlerExecutionQueued
}

func ExecutionQueuedFromExecutionQueue(e *ExecutionQueue) *ExecutionQueued {
	return &ExecutionQueued{
		Event:         e.Event,
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
