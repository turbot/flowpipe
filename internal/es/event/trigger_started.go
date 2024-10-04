package event

import "github.com/turbot/pipe-fittings/modconfig"

type TriggerStarted struct {
	Event   *Event             `json:"event"`
	Trigger *modconfig.Trigger `json:"trigger"`
	Args    modconfig.Input    `json:"args"`
}

func (e *TriggerStarted) GetEvent() *Event {
	return e.Event
}

func (e *TriggerStarted) HandlerName() string {
	return HandlerTriggerStarted
}

func TriggerStartedFromTriggerStart(s *TriggerStart, trigger *modconfig.Trigger) *TriggerStarted {
	return &TriggerStarted{
		Event:   s.Event,
		Args:    s.Args,
		Trigger: trigger,
	}
}
