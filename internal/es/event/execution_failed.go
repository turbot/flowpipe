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
		Event: q.Event,
		Name:  q.Name,
	}
}
