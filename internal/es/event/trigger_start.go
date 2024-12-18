package event

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type TriggerStart struct {
	Event *Event         `json:"event"`
	Name string          `json:"name"`
	Args resources.Input `json:"args"`
}

func (e *TriggerStart) GetEvent() *Event {
	return e.Event
}

func (e *TriggerStart) HandlerName() string {
	return CommandTriggerStart
}

func TriggerStartFromTriggerQueued(q *TriggerQueued) *TriggerStart {
	return &TriggerStart{
		Event: NewFlowEvent(q.Event),
		Name:  q.Name,
		Args:  q.Args,
	}
}
