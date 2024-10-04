package event

type ExecutionFinish struct {
	Event *Event `json:"event"`
	Type  string `json:"type"`
}

func (e *ExecutionFinish) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionFinish) HandlerName() string {
	return CommandExecutionFinish
}

func ExecutionFinishFromExecutionPlanned(e *ExecutionPlanned) *ExecutionFinish {
	return &ExecutionFinish{
		Event: NewFlowEvent(e.Event),
		Type:  e.Type,
	}
}
