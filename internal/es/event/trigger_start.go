package event

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type TriggerStart struct {
	Event *Event         `json:"event"`
	Name  string         `json:"name"`
	Args  flowpipe.Input `json:"args"`
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
