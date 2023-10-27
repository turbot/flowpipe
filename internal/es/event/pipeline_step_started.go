package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
)

// ! This event is only for starting Pipeline Step, not a general step start command.
// !
// ! For general Step Start command, refer to PipelineStepStart
//
// TODO: at one point we should rename this command, make it more clear.
type PipelineStepStarted struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	// Optional details for step execution
	ChildPipelineExecutionID string          `json:"child_pipeline_execution_id,omitempty"`
	ChildPipelineName        string          `json:"child_pipeline_name,omitempty"`
	ChildPipelineArgs        modconfig.Input `json:"child_pipeline_args,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepStartedOption func(*PipelineStepStarted) error

// NewPipelineStepStarted creates a new PipelineStepStarted event.
func NewPipelineStepStarted(opts ...PipelineStepStartedOption) (*PipelineStepStarted, error) {
	// Defaults
	e := &PipelineStepStarted{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineStepStart(cmd *PipelineStepStart) PipelineStepStartedOption {
	return func(e *PipelineStepStarted) error {
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
func WithNewChildPipelineExecutionID() PipelineStepStartedOption {
	return func(e *PipelineStepStarted) error {
		e.ChildPipelineExecutionID = util.NewPipelineExecutionID()
		return nil
	}
}

// WithChildPipelineExecutionID returns a PipelineStepStartedOption that sets
// the ChildPipelineExecutionID to the given ID.
func WithChildPipelineExecutionID(id string) PipelineStepStartedOption {
	return func(e *PipelineStepStarted) error {
		e.ChildPipelineExecutionID = id
		return nil
	}
}

func WithChildPipeline(name string, args modconfig.Input) PipelineStepStartedOption {
	return func(cmd *PipelineStepStarted) error {
		cmd.ChildPipelineName = name
		cmd.ChildPipelineArgs = args
		return nil
	}
}
