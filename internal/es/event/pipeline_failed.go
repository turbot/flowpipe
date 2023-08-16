package event

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

type PipelineFailed struct {
	// Event metadata
	Event *Event `json:"event"`

	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`

	// Error details
	Error *pipeline.StepError `json:"error,omitempty"`

	PipelineOutput map[string]interface{} `json:"pipeline_output"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineFailedOption func(*PipelineFailed) error

// NewPipelineFailed creates a new PipelineFailed event.
// Unlike other events, creating a pipeline failed event cannot have an
// error as an option (because we're already handling errors).
func NewPipelineFailed(ctx context.Context, opts ...PipelineFailedOption) *PipelineFailed {

	logger := fplog.Logger(ctx)

	if logger.TraceLevel != "" {
		stackTrace := string(debug.Stack())
		logger.Info("New pipeline failed event created", "stack_trace", stackTrace)
	}

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
func ForPipelineFail(cmd *PipelineFail, pipelineOutput map[string]interface{}) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = cmd.Error
		e.PipelineOutput = pipelineOutput
		return nil
	}
}

// ForPipelineLoadToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineLoad command.
func ForPipelineLoadToPipelineFailed(cmd *PipelineLoad, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelineFinishToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineFinish command.
func ForPipelineFinishToPipelineFailed(cmd *PipelineFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelineQueueToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineQueue command.
func ForPipelineQueueToPipelineFailed(cmd *PipelineQueue, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelineStartToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineStart command.
func ForPipelineStartToPipelineFailed(cmd *PipelineStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineStepQueueToPipelineFailed(cmd *PipelineStepQueue, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelinePlanToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelinePlan command.
func ForPipelinePlanToPipelineFailed(cmd *PipelinePlan, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		if err != nil {
			e.Error = &pipeline.StepError{
				Message:             err.Error(),
				PipelineExecutionID: cmd.PipelineExecutionID,
			}
		} else {
			e.Error = &pipeline.StepError{
				Message:             "pipeline plan failed but no error is specified",
				PipelineExecutionID: cmd.PipelineExecutionID,
			}
		}

		return nil
	}
}

// ForPipelineCancelToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineCancel command.
func ForPipelineCancelToPipelineFailed(cmd *PipelineCancel, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelinePauseToPipelineFailed(cmd *PipelinePause, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineResumeToPipelineFailed(cmd *PipelineResume, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineStepStartToPipelineFailed(cmd *PipelineStepStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineStepFinishToPipelineFailed(cmd *PipelineStepFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &pipeline.StepError{
			Message:             err.Error(),
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}
