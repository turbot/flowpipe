package event

import (
	flowpipe2 "github.com/turbot/flowpipe/internal/resources"
)

type TriggerStarted struct {
	Event   *Event             `json:"event"`
	Trigger *flowpipe2.Trigger `json:"trigger"`
	Args    flowpipe2.Input    `json:"args"`
}

func (e *TriggerStarted) GetEvent() *Event {
	return e.Event
}

func (e *TriggerStarted) HandlerName() string {
	return HandlerTriggerStarted
}

func TriggerStartedFromTriggerStart(s *TriggerStart, trigger *flowpipe2.Trigger) *TriggerStarted {
	return &TriggerStarted{
		Event:   NewFlowEvent(s.Event),
		Args:    s.Args,
		Trigger: trigger,
	}
}
