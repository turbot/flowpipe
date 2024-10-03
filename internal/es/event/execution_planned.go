package event

type ExecutionPlanned struct {
	// Event metadata
	Event *Event `json:"event"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionPlanned) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionPlanned) HandlerName() string {
	return HandlerExecutionPlanned
}

func ExecutionPlannedFromExecutionPlan(e *ExecutionPlan) *ExecutionPlanned {
	return &ExecutionPlanned{
		Event:         NewFlowEvent(e.Event),
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
