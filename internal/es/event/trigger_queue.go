package event

import "github.com/turbot/pipe-fittings/modconfig"

type TriggerQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline details
	Name string          `json:"name"`
	Args modconfig.Input `json:"args" cty:"args"`
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
