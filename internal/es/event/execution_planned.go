package event

type ExecutionPlanned struct {
	// Event metadata
	Event *Event `json:"event"`
	Type  string `json:"type"`

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
		Type:          e.Type,
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
