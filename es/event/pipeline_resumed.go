package event

type PipelineResumed struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Reason for the cancellation
	Reason string `json:"reason,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineResumedOption func(*PipelineResumed) error

// NewPipelineResumed creates a new PipelineResumed event.
func NewPipelineResumed(opts ...PipelineResumedOption) (*PipelineResumed, error) {
	// Defaults
	e := &PipelineResumed{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineResume(evt *PipelineResume) PipelineResumedOption {
	return func(e *PipelineResumed) error {
		e.Event = NewFlowEvent(evt.Event)
		e.PipelineExecutionID = evt.PipelineExecutionID
		e.Reason = evt.Reason
		return nil
	}
}
