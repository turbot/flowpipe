package event

type ExecutionPaused struct {
	Event         *Event         `json:"event"`
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionPaused) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionPaused) HandlerName() string {
	return HandlerExecutionPaused
}

func ExecutionPausedFromExecutionPlan(e *ExecutionPlan) *ExecutionPaused {
	return &ExecutionPaused{
		Event:         NewFlowEvent(e.Event),
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}
