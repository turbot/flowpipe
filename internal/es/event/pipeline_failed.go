package event

import (
	"context"
	"errors"
	"log"
	"runtime/debug"

	"github.com/turbot/pipe-fittings/perr"

	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
)

type PipelineFailed struct {
	// Event metadata
	Event *Event `json:"event"`

	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`

	// Error details
	Error *modconfig.StepError `json:"error,omitempty"`

	PipelineOutput map[string]interface{} `json:"pipeline_output"`
}

func (e *PipelineFailed) GetEvent() *Event {
	return e.Event
}

func (e *PipelineFailed) HandlerName() string {
	return HandlerPipelineFailed
}

// PipelineFailedOption is a function that modifies an Execution instance.
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

func NewPipelineFailedFromStepForEachPlan(cmd *StepForEachPlan, err error) *PipelineFailed {
	e := &PipelineFailed{}

	var errorModel perr.ErrorModel
	if ok := errors.As(err, &errorModel); !ok {
		errorModel = perr.InternalWithMessage(err.Error())
	}

	e.Event = NewFlowEvent(cmd.Event)
	e.PipelineExecutionID = cmd.PipelineExecutionID
	e.Error = &modconfig.StepError{
		Error:               errorModel,
		PipelineExecutionID: cmd.PipelineExecutionID,
	}

	return e
}

func NewPipelineFailedFromPipelineLoad(cmd *PipelineLoad, err error) *PipelineFailed {

	e := &PipelineFailed{}
	var errorModel perr.ErrorModel
	if ok := errors.As(err, &errorModel); !ok {
		errorModel = perr.InternalWithMessage(err.Error())
	}
	e.Event = NewFlowEvent(cmd.Event)
	e.PipelineExecutionID = cmd.PipelineExecutionID
	e.Error = &modconfig.StepError{
		Error:               errorModel,
		PipelineExecutionID: cmd.PipelineExecutionID,
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

// ForPipelineFinishToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineFinish command.
func ForPipelineFinishToPipelineFailed(cmd *PipelineFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelineQueueToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineQueue command.
func ForPipelineQueueToPipelineFailed(cmd *PipelineQueue, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

// ForPipelineStartToPipelineFailed returns a PipelineFailedOption that sets the fields of the
// PipelineFailed event from a PipelineStart command.
func ForPipelineStartToPipelineFailed(cmd *PipelineStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForStepQueueToPipelineFailed(cmd *StepQueue, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
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
			var errorModel perr.ErrorModel
			if ok := errors.As(err, &errorModel); !ok {
				errorModel = perr.InternalWithMessage(err.Error())
			}
			e.Error = &modconfig.StepError{
				Error:               errorModel,
				PipelineExecutionID: cmd.PipelineExecutionID,
			}
		} else {
			e.Error = &modconfig.StepError{
				Error:               perr.InternalWithMessage("pipeline plan failed but no error is specified"),
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
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelinePauseToPipelineFailed(cmd *PipelinePause, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineResumeToPipelineFailed(cmd *PipelineResume, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForStepStartToPipelineFailed(cmd *StepStart, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}

func ForPipelineStepFinishToPipelineFailed(cmd *StepPipelineFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Error = &modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		}
		return nil
	}
}
