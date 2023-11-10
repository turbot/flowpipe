package event

type PipelineCanceled struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Reason for the cancellation
	Reason string `json:"reason,omitempty"`
}

func (e *PipelineCanceled) GetEvent() *Event {
	return e.Event
}

func (e *PipelineCanceled) HandlerName() string {
	return HandlerPipelineCancelled
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineCanceledOption func(*PipelineCanceled) error

// NewPipelineCanceled creates a new PipelineCanceled event.
func NewPipelineCanceled(opts ...PipelineCanceledOption) (*PipelineCanceled, error) {
	// Defaults
	e := &PipelineCanceled{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineCancel returns a PipelineCanceledOption that sets the fields of the
// PipelineCanceled event from a PipelineCancel command.
func ForPipelineCancel(cmd *PipelineCancel) PipelineCanceledOption {
	return func(e *PipelineCanceled) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Reason = cmd.Reason
		return nil
	}
}
