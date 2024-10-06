package event

import "github.com/turbot/pipe-fittings/modconfig"

type TriggerQueued struct {
	Event *Event          `json:"event"`
	Name  string          `json:"name"`
	Args  modconfig.Input `json:"args"`
}

func (e *TriggerQueued) GetEvent() *Event {
	return e.Event
}

func (e *TriggerQueued) HandlerName() string {
	return HandlerTriggerQueued
}

func TriggerQueuedFromTriggerQueue(q *TriggerQueue) *TriggerQueued {
	return &TriggerQueued{
		Event: NewFlowEvent(q.Event),
		Name:  q.Name,
		Args:  q.Args,
	}
}
