package event

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type TriggerStarted struct {
	Event   *Event            `json:"event"`
	Trigger *flowpipe.Trigger `json:"trigger"`
	Args    flowpipe.Input    `json:"args"`
}

func (e *TriggerStarted) GetEvent() *Event {
	return e.Event
}

func (e *TriggerStarted) HandlerName() string {
	return HandlerTriggerStarted
}

func TriggerStartedFromTriggerStart(s *TriggerStart, trigger *flowpipe.Trigger) *TriggerStarted {
	return &TriggerStarted{
		Event:   NewFlowEvent(s.Event),
		Args:    s.Args,
		Trigger: trigger,
	}
}
