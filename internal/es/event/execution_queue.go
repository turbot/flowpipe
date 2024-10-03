package event

type ExecutionQueue struct {
	Event *Event `json:"event"`
	Type  string `json:"type"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionQueue) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionQueue) HandlerName() string {
	return CommandExecutionQueue
}
