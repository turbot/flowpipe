package event

import "github.com/turbot/pipe-fittings/perr"

type ExecutionFail struct {
	Event *Event          `json:"event"`
	Name  string          `json:"name"`
	Error perr.ErrorModel `json:"error"`
}

func (e *ExecutionFail) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionFail) HandlerName() string {
	return CommandExecutionFail
}

func ExecutionFailFromTriggerFailed(q *TriggerFailed) *ExecutionFail {
	return &ExecutionFail{
		Event: NewFlowEvent(q.Event),
		Name:  q.Name,
	}
}

func ExecutionFailFromTriggerStarted(q *TriggerStarted, err perr.ErrorModel) *ExecutionFail {
	return &ExecutionFail{
		Event: NewFlowEvent(q.Event),
		Name:  q.Trigger.Name(),
		Error: err,
	}
}
