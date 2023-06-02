package event

import "log"

type PipelineFailed struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Error details
	ErrorMessage string `json:"error_message"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineFailedOption func(*PipelineFailed) error

// NewPipelineFailed creates a new PipelineFailed event.
// Unlike other events, creating a pipeline failed event cannot have an
// error as an option (because we're already handling errors).
func NewPipelineFailed(opts ...PipelineFailedOption) *PipelineFailed {
	// Defaults
	e := &PipelineFailed{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			log.Fatalf("error creating pipeline failed event: %v", err)
		}
	}
	return e
}

// ForPipelineFail returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineFail command.
func ForPipelineFail(cmd *PipelineFail) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = cmd.ErrorMessage
		return nil
	}
}

// ForPipelineLoadToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineLoad command.
func ForPipelineLoadToPipelineFailed(cmd *PipelineLoad, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

// ForPipelineFinishToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineFinish command.
func ForPipelineFinishToPipelineFailed(cmd *PipelineFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

// ForPipelineQueueToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineQueue command.
func ForPipelineQueueToPipelineFailed(cmd *PipelineQueue, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

// ForPipelineStartToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineStart command.
func ForPipelineStartToPipelineFailed(cmd *PipelineStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

// ForPipelinePlanToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelinePlan command.
func ForPipelinePlanToPipelineFailed(cmd *PipelinePlan, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

// ForPipelineCancelToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineCancel command.
func ForPipelineCancelToPipelineFailed(cmd *PipelineCancel, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

func ForPipelineStepStartToPipelineFailed(cmd *PipelineStepStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}

func ForPipelineStepFinishToPipelineFailed(cmd *PipelineStepFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.ErrorMessage = err.Error()
		return nil
	}
}
