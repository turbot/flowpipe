package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStepStart struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string          `json:"pipeline_execution_id"`
	StepExecutionID     string          `json:"step_execution_id"`
	StepName            string          `json:"step_name"`
	StepInput           modconfig.Input `json:"input"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`

	DelayMs        int                      `json:"delay_ms,omitempty"` // delay start in milliseconds
	NextStepAction modconfig.NextStepAction `json:"next_step_action,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepStartOption func(*PipelineStepStart) error

// NewPipelineStepStart creates a new PipelineStepStart event.
func NewPipelineStepStart(opts ...PipelineStepStartOption) (*PipelineStepStart, error) {
	// Defaults
	e := &PipelineStepStart{
		StepExecutionID: util.NewStepExecutionID(),
	}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelinePlanned(e *PipelinePlanned) PipelineStepStartOption {
	return func(cmd *PipelineStepStart) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		return nil
	}
}

func ForPipelineStepQueued(e *PipelineStepQueued) PipelineStepStartOption {
	return func(cmd *PipelineStepStart) error {

		if e.StepExecutionID == "" {
			return perr.BadRequestWithMessage("missing step execution ID in pipeline step queued event")
		}

		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		cmd.StepExecutionID = e.StepExecutionID
		return nil
	}
}

func WithStep(name string, input modconfig.Input, stepForEach *modconfig.StepForEach, nextStepAction modconfig.NextStepAction) PipelineStepStartOption {
	return func(cmd *PipelineStepStart) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.StepForEach = stepForEach
		cmd.NextStepAction = nextStepAction
		return nil
	}
}
