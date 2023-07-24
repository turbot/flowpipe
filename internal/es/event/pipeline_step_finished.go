package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/types"
)

type PipelineStepFinished struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`

	// Output
	Output *types.StepOutput `json:"output,omitempty"`

	// for_each controls
	StepForEach *types.StepForEach `json:"step_for_each,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepFinishedOption func(*PipelineStepFinished) error

// NewPipelineStepFinished creates a new PipelineStepFinished event.
func NewPipelineStepFinished(opts ...PipelineStepFinishedOption) (*PipelineStepFinished, error) {
	// Defaults
	e := &PipelineStepFinished{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineStepStartToPipelineStepFinished(cmd *PipelineStepStart) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Event = NewFlowEvent(cmd.Event)
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
		e.StepForEach = cmd.StepForEach
		return nil
	}
}

func ForPipelineStepFinish(cmd *PipelineStepFinish) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline step finish command: %v", e)
		}
		if cmd.StepExecutionID != "" {
			e.StepExecutionID = cmd.StepExecutionID
		} else {
			return fmt.Errorf("missing step execution ID in pipeline step finish command: %v", e)
		}
		e.Output = cmd.Output
		e.StepForEach = cmd.StepForEach
		return nil
	}
}

func WithStepOutput(output *types.StepOutput) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Output = output
		return nil
	}
}
