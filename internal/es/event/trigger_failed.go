package event

type TriggerFailed struct {
	Event *Event `json:"event"`
	Name  string `json:"name"`
}

func (e *TriggerFailed) GetEvent() *Event {
	return e.Event
}

func (e *TriggerFailed) HandlerName() string {
	return HandlerTriggerFailed
}

func TriggerFailedFromTriggerStart(q *TriggerStart) *TriggerFailed {
	return &TriggerFailed{
		Event: NewFlowEvent(q.Event),
		Name:  q.Name,
	}
}

func TriggerFailedFromExecutionPlan(q *ExecutionPlan, triggerName string) *TriggerFailed {
	return &TriggerFailed{
		Event: NewFlowEvent(q.Event),
		Name:  triggerName,
	}
}
