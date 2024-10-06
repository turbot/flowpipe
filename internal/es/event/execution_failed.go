package event

type ExecutionFailed struct {
	Event *Event `json:"event"`
	Name  string `json:"name"`
}

func (e *ExecutionFailed) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionFailed) HandlerName() string {
	return HandlerExecutionFailed
}

func ExecutionFailedFromExecutionFail(q *ExecutionFail) *ExecutionFailed {
	return &ExecutionFailed{
		Event: NewFlowEvent(q.Event),
		Name:  q.Name,
	}
}

func ExecutionFailedFromExecutionPlan(q *ExecutionPlan) *ExecutionFailed {
	return &ExecutionFailed{
		Event: NewFlowEvent(q.Event),
		Name:  q.Type,
	}
}
