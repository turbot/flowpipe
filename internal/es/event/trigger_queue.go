package event

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type TriggerQueue struct {
	Event              *Event         `json:"event"`
	TriggerExecutionID string         `json:"trigger_execution_id"`
	Name string          `json:"name"`
	Args resources.Input `json:"args"`
}

func (e *TriggerQueue) GetEvent() *Event {
	return e.Event
}

func (e *TriggerQueue) HandlerName() string {
	return CommandTriggerQueue
}

func (e *TriggerQueue) GetName() string {
	return e.Name
}

func (e *TriggerQueue) GetType() string {
	return "trigger"
}
