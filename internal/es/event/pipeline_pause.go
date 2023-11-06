package event

type PipelinePause struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	ExecutionID         string `json:"execution_id,omitempty"`

	// Reason for pausing the pipeline execution
	Reason string `json:"reason,omitempty"`
}

func (e *PipelinePause) GetEvent() *Event {
	return e.Event
}

func (e *PipelinePause) HandlerName() string {
	return "command.pipeline_pause"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelinePauseOption func(*PipelinePause) error

// NewPipelineCancel creates a new PipelineCancel event.
func NewPipelinePause(pipelineExecutionID string, opts ...PipelinePauseOption) (*PipelinePause, error) {
	// Defaults
	e := NewEventForExecutionID(pipelineExecutionID)
	// Defaults
	cmd := &PipelinePause{
		Event: e,
	}
	// Set options
	for _, opt := range opts {
		err := opt(cmd)
		if err != nil {
			return cmd, err
		}
	}
	return cmd, nil
}
