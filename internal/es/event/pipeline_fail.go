package event

import "github.com/turbot/flowpipe/pipeparser/pipeline"

type PipelineFail struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Error details
	Error *pipeline.StepError `json:"error,omitempty"`
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
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineQueuedToPipelineFail(e *PipelineQueued, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineStartedToPipelineFail(e *PipelineStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineResumedToPipelineFail(e *PipelineResumed, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineStepStartedToPipelineFail(e *PipelineStepStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineStepFinishedToPipelineFail(e *PipelineStepFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		if err != nil {
			cmd.Error = &pipeline.StepError{
				Message:             err.Error(),
				PipelineExecutionID: e.PipelineExecutionID,
			}
		}
	}
}

func ForPipelinePlannedToPipelineFail(e *PipelinePlanned, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		if err != nil {
			cmd.Error = &pipeline.StepError{
				Message:             err.Error(),
				PipelineExecutionID: e.PipelineExecutionID,
			}
		}
	}
}

func ForPipelineStepQueuedToPipelineFail(e *PipelineStepQueued, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineFinishedToPipelineFail(e *PipelineFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}
