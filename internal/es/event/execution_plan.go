package event

type ExecutionPlan struct {
	Event *Event `json:"event"`
	Type  string `json:"type"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionPlan) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionPlan) HandlerName() string {
	return CommandExecutionPlan
}

func ExecutionPlanFromExecutionStarted(e *ExecutionStarted) *ExecutionPlan {
	return &ExecutionPlan{
		Event:         NewFlowEvent(e.Event),
		PipelineQueue: e.PipelineQueue,
		TriggerQueue:  e.TriggerQueue,
	}
}

func ExecutionPlanFromPipelineFinished(e *PipelineFinished) *ExecutionPlan {
	return &ExecutionPlan{
		Event: NewFlowEvent(e.Event),
	}
}

func ExecutionPlanFromPipelineFailed(e *PipelineFailed) *ExecutionPlan {
	return &ExecutionPlan{
		Event: NewFlowEvent(e.Event),
	}
}

func ExecutionPlanFromPipelineCancelled(e *PipelineCanceled) *ExecutionPlan {
	return &ExecutionPlan{
		Event: NewFlowEvent(e.Event),
	}
}

func ExecutionPlanFromPipelinePaused(e *PipelinePaused) *ExecutionPlan {
	return &ExecutionPlan{
		Event: NewFlowEvent(e.Event),
	}
}
