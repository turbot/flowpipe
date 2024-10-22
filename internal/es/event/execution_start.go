package event

type ExecutionStart struct {
	// Event metadata
	Event *Event `json:"event"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"` // PipelineQueue not supported yet

	TriggerQueue *TriggerQueue `json:"trigger_queue"`
}

func (e *ExecutionStart) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionStart) HandlerName() string {
	return CommandExecutionStart
}

func ExecutionStartFromExecutionQueued(e *ExecutionQueued) *ExecutionStart {
	return &ExecutionStart{
		Event:         NewFlowEvent(e.Event),
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
