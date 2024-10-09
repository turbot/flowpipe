package event

type TriggerFinished struct {
	Event *Event `json:"event"`
	Name  string `json:"name"`
}

func (e *TriggerFinished) GetEvent() *Event {
	return e.Event
}

func (e *TriggerFinished) HandlerName() string {
	return HandlerTriggerFinished
}

func TriggerFinishedFromExecutionPlan(q *ExecutionPlan, triggerName string) *TriggerFinished {
	return &TriggerFinished{
		Event: NewFlowEvent(q.Event),
		Name:  triggerName,
	}
}

func TriggerFinishedFromTriggerFinish(cmd *TriggerFinish) *TriggerFinished {
	return &TriggerFinished{
		Event: cmd.Event,
		Name:  cmd.Name,
	}
}
