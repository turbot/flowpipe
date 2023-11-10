package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
)

// ! This event is only for starting Pipeline Step, not a general step start command.
// !
// ! For general Step Start command, refer to StepStart
type StepPipelineStarted struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	// Optional details for step execution
	ChildPipelineExecutionID string          `json:"child_pipeline_execution_id,omitempty"`
	ChildPipelineName        string          `json:"child_pipeline_name,omitempty"`
	ChildPipelineArgs        modconfig.Input `json:"child_pipeline_args,omitempty"`

	// The key for a single step is always "0" but say this pipeline step start is part of for_each, the key is
	// populated with the actual key: "0"/"1"/"2" or "foo"/"bar"/"baz" (for map based for_each)
	//
	// This key is only relevant to its immediate parent (if we have multiple nested pipelines)
	Key string `json:"key"`
}

func (e *StepPipelineStarted) GetEvent() *Event {
	return e.Event
}

func (e *StepPipelineStarted) HandlerName() string {
	return "handler.step_pipeline_started"
}

type StepPipelineStartedOption func(*StepPipelineStarted) error

// NewStepPipelineStarted creates a new StepPipelineStarted event.
func NewStepPipelineStarted(opts ...StepPipelineStartedOption) (*StepPipelineStarted, error) {
	// Defaults
	e := &StepPipelineStarted{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForStepStart(cmd *StepStart) StepPipelineStartedOption {
	return func(e *StepPipelineStarted) error {
		e.Event = NewChildEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline step start command: %v", e)
		}
		if cmd.StepExecutionID != "" {
			e.StepExecutionID = cmd.StepExecutionID
		} else {
			return fmt.Errorf("missing step execution ID in pipeline step start command: %v", e)
		}
		return nil
	}
}

// WithNewChildPipelineExecutionID returns a PipelineStepStartedOption that sets
// the ChildPipelineExecutionID to a new ID.
func WithNewChildPipelineExecutionID() StepPipelineStartedOption {
	return func(e *StepPipelineStarted) error {
		e.ChildPipelineExecutionID = util.NewPipelineExecutionID()
		return nil
	}
}

// WithChildPipelineExecutionID returns a PipelineStepStartedOption that sets
// the ChildPipelineExecutionID to the given ID.
func WithChildPipelineExecutionID(id string) StepPipelineStartedOption {
	return func(e *StepPipelineStarted) error {
		e.ChildPipelineExecutionID = id
		return nil
	}
}

func WithChildPipeline(name string, args modconfig.Input) StepPipelineStartedOption {
	return func(cmd *StepPipelineStarted) error {
		cmd.ChildPipelineName = name
		cmd.ChildPipelineArgs = args
		return nil
	}
}
