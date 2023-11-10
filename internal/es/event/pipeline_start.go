package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStart struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
}

func (e *PipelineStart) GetEvent() *Event {
	return e.Event
}

func (e *PipelineStart) HandlerName() string {
	return CommandPipelineStart
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStartOption func(*PipelineStart) error

// NewPipelineStart creates a new PipelineStart event.
func NewPipelineStart(opts ...PipelineStartOption) (*PipelineStart, error) {
	// Defaults
	e := &PipelineStart{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineStart returns a PipelineStartOption that sets the fields of the
// PipelineStart event from a PipelineStart command.
func ForPipelineLoaded(e *PipelineLoaded) PipelineStartOption {
	return func(cmd *PipelineStart) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline loaded event: %v", e))
		}
		return nil
	}
}
