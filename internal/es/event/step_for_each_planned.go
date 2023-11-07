package event

type StepForEachPlanned struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline details
	Name string `json:"name"`
}

func (e *StepForEachPlanned) GetEvent() *Event {
	return e.Event
}

func (e *StepForEachPlanned) HandlerName() string {
	return "handler.step_for_each_planned"
}
