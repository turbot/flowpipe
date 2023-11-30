package event

import (
	"context"
	"encoding/json"
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

	Errors []modconfig.StepError `json:"error,omitempty"`

	PipelineOutput map[string]interface{} `json:"pipeline_output"`
}

func (e *PipelineFailed) GetEvent() *Event {
	return e.Event
}

func (e *PipelineFailed) HandlerName() string {
	return HandlerPipelineFailed
}

func (p *PipelineFailed) UnmarshalJSON(data []byte) error {
	type Alias PipelineFailed
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Custom handling for PipelineOutput
	rawErrors, ok := aux.PipelineOutput["errors"]
	if !ok {
		return nil
	}

	var stepErrors []modconfig.StepError
	rawErrorsBytes, err := json.Marshal(rawErrors)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawErrorsBytes, &stepErrors); err != nil {
		return err
	}

	// Update the PipelineOutput with the strongly typed errors
	p.PipelineOutput["errors"] = stepErrors

	return nil
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
	e.Errors = []modconfig.StepError{
		{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
	e.Errors = []modconfig.StepError{
		{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
	}
	return e
}

func NewPipelineFailedFromPipelineFail(cmd *PipelineFail, pipelineOutput map[string]interface{}, pipelineErrors []modconfig.StepError) *PipelineFailed {
	e := &PipelineFailed{}
	e.Event = NewFlowEvent(cmd.Event)
	e.PipelineExecutionID = cmd.PipelineExecutionID
	e.PipelineOutput = pipelineOutput
	e.Errors = pipelineErrors
	return e
}

func NewPipelineFailedFromStepPipelineFinish(cmd *StepPipelineFinish, err error) *PipelineFailed {
	e := &PipelineFailed{}

	var errorModel perr.ErrorModel
	if ok := errors.As(err, &errorModel); !ok {
		errorModel = perr.InternalWithMessage(err.Error())
	}

	e.Event = NewFlowEvent(cmd.Event)
	e.PipelineExecutionID = cmd.PipelineExecutionID
	stepError := modconfig.StepError{
		Error:               errorModel,
		PipelineExecutionID: cmd.PipelineExecutionID,
		StepExecutionID:     cmd.StepExecutionID,
	}

	e.Errors = append(e.Errors, stepError)
	return e
}

func ForPipelineFinishToPipelineFailed(cmd *PipelineFinish, err error) PipelineFailedOption {
	return func(e *PipelineFailed) error {
		var errorModel perr.ErrorModel
		if ok := errors.As(err, &errorModel); !ok {
			errorModel = perr.InternalWithMessage(err.Error())
		}
		e.Event = NewFlowEvent(cmd.Event)
		e.PipelineExecutionID = cmd.PipelineExecutionID
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
			e.Errors = []modconfig.StepError{{
				Error:               errorModel,
				PipelineExecutionID: cmd.PipelineExecutionID,
			},
			}
		} else {
			e.Errors = []modconfig.StepError{{
				Error:               perr.InternalWithMessage("pipeline plan failed but no error is specified"),
				PipelineExecutionID: cmd.PipelineExecutionID,
			},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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
		e.Errors = []modconfig.StepError{{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
		},
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

		stepError := modconfig.StepError{
			Error:               errorModel,
			PipelineExecutionID: cmd.PipelineExecutionID,
			StepExecutionID:     cmd.StepExecutionID,
			Step:                cmd.StepName,
		}

		if e.PipelineOutput == nil {
			e.PipelineOutput = map[string]interface{}{
				"errors": []modconfig.StepError{
					stepError,
				},
			}
		} else {
			e.PipelineOutput["errors"] = append(e.PipelineOutput["errors"].([]modconfig.StepError), stepError)
		}
		return nil
	}
}
