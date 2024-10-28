package event

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type TriggerQueued struct {
	Event              *Event         `json:"event"`
	TriggerExecutionID string         `json:"trigger_execution_id"`
	Name               string         `json:"name"`
	Args               flowpipe.Input `json:"args"`
}

func (e *TriggerQueued) GetEvent() *Event {
	return e.Event
}

func (e *TriggerQueued) HandlerName() string {
	return HandlerTriggerQueued
}

func TriggerQueuedFromTriggerQueue(q *TriggerQueue) *TriggerQueued {
	return &TriggerQueued{
		Event:              NewFlowEvent(q.Event),
		TriggerExecutionID: q.TriggerExecutionID,
		Name:               q.Name,
		Args:               q.Args,
	}
}
