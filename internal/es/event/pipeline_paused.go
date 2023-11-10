package event

type PipelinePaused struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	//Reason for pausing the pipeline execution
	Reason string `json:"reason,omitempty"`
}

func (e *PipelinePaused) GetEvent() *Event {
	return e.Event
}

func (e *PipelinePaused) HandlerName() string {
	return HandlerPipelinePaused
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelinePausedOption func(*PipelinePaused) error

// NewPipelineCancel creates a new PipelineCancel event.
func NewPipelinePaused(opts ...PipelinePausedOption) (*PipelinePaused, error) {
	// Defaults
	e := &PipelinePaused{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelinePause(cmd *PipelinePause) PipelinePausedOption {
	return func(e *PipelinePaused) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Reason = cmd.Reason
		return nil
	}
}
