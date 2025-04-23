package event

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type TriggerFinish struct {
	Event              *Event `json:"event"`
	TriggerExecutionID string
	Name string          `json:"name"`
	Args resources.Input `json:"args"`
}

func (e *TriggerFinish) GetEvent() *Event {
	return e.Event
}

func (e *TriggerFinish) HandlerName() string {
	return CommandTriggerFinish
}

func (e *TriggerFinish) GetName() string {
	return e.Name
}

func (e *TriggerFinish) GetType() string {
	return "trigger"
}

func TrigerFinishFromTriggerStarted(q *TriggerStarted) *TriggerFinish {
	return &TriggerFinish{
		Event: NewFlowEvent(q.Event),
		Name:  q.Trigger.Name(),
		Args:  q.Args,
	}
}
