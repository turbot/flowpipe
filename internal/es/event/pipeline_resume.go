package event

type PipelineResume struct {
	// Event metadata
	Event *Event `json:"event"`

	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	ExecutionID         string `json:"execution_id,omitempty"`

	// Reason for the cancellation
	Reason string `json:"reason,omitempty"`
}

func (e *PipelineResume) GetEvent() *Event {
	return e.Event
}

func (e *PipelineResume) HandlerName() string {
	return CommandPipelineResume
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineResumeOption func(*PipelineResume) error

func NewPipelineResume(executionId, pipelineExecutionId string) *PipelineResume {
	// Defaults
	e := NewEventForExecutionID(executionId)
	// Defaults
	evt := &PipelineResume{
		Event:               e,
		PipelineExecutionID: pipelineExecutionId,
	}
	return evt
}
