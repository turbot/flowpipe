package event

type ExecutionStarted struct {
	// Event metadata
	Event *Event `json:"event"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionStarted) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionStarted) HandlerName() string {
	return HandlerExecutionStarted
}

func ExecutionStartedFromExecutionStart(e *ExecutionStart) *ExecutionStarted {
	return &ExecutionStarted{
		Event:         NewFlowEvent(e.Event),
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
