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

func NewPipelineCanceledFromPipelineCancel(cmd *PipelineCancel) *PipelineCanceled {
	e := &PipelineCanceled{
		Event:               NewFlowEvent(cmd.Event),
		PipelineExecutionID: cmd.PipelineExecutionID,
		Reason:              cmd.Reason,
	}
	return e
}
