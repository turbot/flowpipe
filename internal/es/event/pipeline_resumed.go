package event

type PipelineResumed struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Reason for the cancellation
	Reason string `json:"reason,omitempty"`
}

func (e *PipelineResumed) GetEvent() *Event {
	return e.Event
}

func (e *PipelineResumed) HandlerName() string {
	return HandlerPipelineResumed
}

// NewPipelineResumed creates a new PipelineResumed event.
func NewPipelineResumedFromPipelineResume(evt *PipelineResume) *PipelineResumed {
	e := &PipelineResumed{
		Event:               NewFlowEvent(evt.Event),
		PipelineExecutionID: evt.PipelineExecutionID,
		Reason:              evt.Reason,
	}
	return e
}
