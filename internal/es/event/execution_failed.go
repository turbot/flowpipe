package event

import "github.com/turbot/pipe-fittings/perr"

type ExecutionFailed struct {
	Event *Event          `json:"event"`
	Name  string          `json:"name"`
	Error perr.ErrorModel `json:"error"`
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
		Error: q.Error,
	}
}

func ExecutionFailedFromExecutionPlan(q *ExecutionPlan, err perr.ErrorModel) *ExecutionFailed {
	return &ExecutionFailed{
		Event: NewFlowEvent(q.Event),
		Name:  q.Type,
		Error: err,
	}
}
