package event

import "github.com/turbot/pipe-fittings/modconfig"

type TriggerQueued struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline details
	Name string          `json:"name"`
	Args modconfig.Input `json:"args" cty:"args"`
}

func (e *TriggerQueued) GetEvent() *Event {
	return e.Event
}

func (e *TriggerQueued) HandlerName() string {
	return HandlerTriggerQueued
}
