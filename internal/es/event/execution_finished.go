package event

type ExecutionFinished struct {
	Event *Event `json:"event"`
	Type  string `json:"type"`
}

func (e *ExecutionFinished) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionFinished) HandlerName() string {
	return HandlerExecutionFinished
}

func ExecutionFinishedFromExecutionFinish(e *ExecutionFinish) *ExecutionFinished {
	return &ExecutionFinished{
		Event: NewFlowEvent(e.Event),
		Type:  e.Type,
	}
}
