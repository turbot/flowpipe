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
	return CommandPipelinePause
}

type PipelinePauseOption func(*PipelinePause) error

func NewPipelinePause(executionId, pipelineExecutionId string, opts ...PipelinePauseOption) (*PipelinePause, error) {
	// Defaults
	e := NewEventForExecutionID(executionId)
	// Defaults
	cmd := &PipelinePause{
		Event:               e,
		PipelineExecutionID: pipelineExecutionId,
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

func PipelinePauseFromPipelinePlanned(e *PipelinePlanned) *PipelinePause {
	cmd := &PipelinePause{
		Event:               NewFlowEvent(e.Event),
		PipelineExecutionID: e.PipelineExecutionID,
	}
	return cmd
}
