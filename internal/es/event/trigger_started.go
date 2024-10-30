package event

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type TriggerStarted struct {
	Event   *Event             `json:"event"`
	Trigger *resources.Trigger `json:"trigger"`
	Args    resources.Input    `json:"args"`
}

func (e *TriggerStarted) GetEvent() *Event {
	return e.Event
}

func (e *TriggerStarted) HandlerName() string {
	return HandlerTriggerStarted
}

func TriggerStartedFromTriggerStart(s *TriggerStart, trigger *resources.Trigger) *TriggerStarted {
	return &TriggerStarted{
		Event:   NewFlowEvent(s.Event),
		Args:    s.Args,
		Trigger: trigger,
	}
}
