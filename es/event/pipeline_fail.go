package event

type PipelineFail struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Error details
	ErrorMessage string `json:"error_message"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineFailOption func(*PipelineFail)

// NewPipelineFail creates a new PipelineFail event.
// Unlike other events, creating a pipeline fail event cannot have an
// error as an option (because we're already handling errors).
func NewPipelineFail(opts ...PipelineFailOption) *PipelineFail {
	// Defaults
	cmd := &PipelineFail{}
	// Set options
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func ForPipelineLoadedToPipelineFail(e *PipelineLoaded, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineQueuedToPipelineFail(e *PipelineQueued, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineStartedToPipelineFail(e *PipelineStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineResumedToPipelineFail(e *PipelineResumed, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineStepStartedToPipelineFail(e *PipelineStepStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineStepFinishedToPipelineFail(e *PipelineStepFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelinePlannedToPipelineFail(e *PipelinePlanned, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}

func ForPipelineFinishedToPipelineFail(e *PipelineFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.ErrorMessage = err.Error()
	}
}
