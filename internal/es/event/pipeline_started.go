package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStarted struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStartedOption func(*PipelineStarted) error

// NewPipelineStarted creates a new PipelineStarted event.
func NewPipelineStarted(opts ...PipelineStartedOption) (*PipelineStarted, error) {
	// Defaults
	e := &PipelineStarted{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineStart returns a PipelineStartedOption that sets the fields of the
// PipelineStarted event from a PipelineStart command.
func ForPipelineStart(cmd *PipelineStart) PipelineStartedOption {
	return func(e *PipelineStarted) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline start command: %v", cmd))
		}
		return nil
	}
}
