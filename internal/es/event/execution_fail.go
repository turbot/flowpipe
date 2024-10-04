package event

type ExecutionFail struct {
	Event *Event `json:"event"`
	Name  string `json:"name"`
	Error error  `json:"error"`
}

func (e *ExecutionFail) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionFail) HandlerName() string {
	return CommandExecutionFail
}

func ExecutionFailFromTriggerFailed(q *TriggerFailed) *ExecutionFail {
	return &ExecutionFail{
		Event: q.Event,
		Name:  q.Name,
	}
}

func ExecutionFailFromTriggerStarted(q *TriggerStarted, err error) *ExecutionFail {
	return &ExecutionFail{
		Event: q.Event,
		Name:  q.Trigger.Name(),
		Error: err,
	}
}
