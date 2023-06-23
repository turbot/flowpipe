package event

type PipelineCancel struct {
	// Event metadata
	Event *Event `json:"event"`

	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	ExecutionID         string `json:"execution_id,omitempty"`

	// Reason for the cancellation
	Reason string `json:"reason,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineCancelOption func(*PipelineCancel) error

// NewPipelineCancel creates a new PipelineCancel event.
func NewPipelineCancel(pipelineExecutionID string, opts ...PipelineCancelOption) (*PipelineCancel, error) {
	// Defaults
	e := NewEventForExecutionID(pipelineExecutionID)
	// Defaults
	evt := &PipelineCancel{
		Event: e,
	}
	// Set options
	for _, opt := range opts {
		err := opt(evt)
		if err != nil {
			return evt, err
		}
	}
	return evt, nil
}
