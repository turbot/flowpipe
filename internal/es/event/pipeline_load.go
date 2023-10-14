package event

import (
	"fmt"

	"github.com/turbot/flowpipe/pipeparser/perr"
)

type PipelineLoad struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineLoadOption func(*PipelineLoad) error

// NewPipelineLoad creates a new PipelineLoad event.
func NewPipelineLoad(opts ...PipelineLoadOption) (*PipelineLoad, error) {
	// Defaults
	e := &PipelineLoad{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineLoad returns a PipelineLoadOption that sets the fields of the
// PipelineLoad event from a PipelineLoad command.
func ForPipelineQueued(e *PipelineQueued) PipelineLoadOption {
	return func(cmd *PipelineLoad) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline queued event: %v", e))
		}
		return nil
	}
}
