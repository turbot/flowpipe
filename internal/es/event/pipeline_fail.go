package event

import (
	"errors"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineFail struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Error details
	Error *resources.StepError `json:"error,omitempty"`
}

func (e *PipelineFail) GetEvent() *Event {
	return e.Event
}

func (e *PipelineFail) HandlerName() string {
	return CommandPipelineFail
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

func NewPipelineFailFromPipelinePlanned(e *PipelinePlanned, err error) *PipelineFail {
	cmd := PipelineFail{}
	cmd.Event = NewFlowEvent(e.Event)
	cmd.PipelineExecutionID = e.PipelineExecutionID
	if err != nil {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
	return &cmd
}

func PipelineFailFromPipelinePaused(e *PipelinePaused, err error) *PipelineFail {
	cmd := PipelineFail{
		Event:               NewFlowEvent(e.Event),
		PipelineExecutionID: e.PipelineExecutionID,
	}

	if err != nil {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
	return &cmd
}

func NewPipelineFailFromStepForEachPlanned(e *StepForEachPlanned, err error) *PipelineFail {
	cmd := PipelineFail{}
	cmd.Event = NewFlowEvent(e.Event)
	cmd.PipelineExecutionID = e.PipelineExecutionID
	if err != nil {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
	return &cmd
}

func ForPipelineLoadedToPipelineFail(e *PipelineLoaded, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineQueuedToPipelineFail(e *PipelineQueued, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)

		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}

		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineStartedToPipelineFail(e *PipelineStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)

		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineResumedToPipelineFail(e *PipelineResumed, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)

		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}

		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForStepPipelineStartedToPipelineFail(e *StepPipelineStarted, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)

		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}

		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineStepFinishedToPipelineFail(e *StepFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		var errorModel perr.ErrorModel

		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID

		if err != nil {
			if ok := errors.As(err, &errorModel); !ok {
				errorModel = perr.InternalWithMessage(err.Error())
			}
			cmd.Error = &resources.StepError{
				Error:               errorModel,
				PipelineExecutionID: e.PipelineExecutionID,
			}
		}
	}
}

func ForStepQueuedToPipelineFail(e *StepQueued, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID
		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}

func ForPipelineFinishedToPipelineFail(e *PipelineFinished, err error) PipelineFailOption {
	return func(cmd *PipelineFail) {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = e.PipelineExecutionID

		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}

		cmd.Error = &resources.StepError{
			Error:               errorModel,
			PipelineExecutionID: e.PipelineExecutionID,
		}
	}
}
