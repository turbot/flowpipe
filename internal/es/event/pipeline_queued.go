package event

import (
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
)

// PipelineQueued is published when a pipeline is queued
type PipelineQueued struct {
	// Event metadata
	Event *Event `json:"event"`
	// Name of the pipeline to be queued
	Name string `json:"name"`
	// Input to the pipeline
	Args modconfig.Input `json:"args"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// If this is a child pipeline then set the parent step execution ID
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
	ParentExecutionID     string `json:"parent_execution_id,omitempty"`
}

func (e *PipelineQueued) GetEvent() *Event {
	return e.Event
}

func (e *PipelineQueued) HandlerName() string {
	return HandlerPipelineQueued
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineQueuedOption func(*PipelineQueued) error

// NewPipelineQueued creates a new PipelineQueued event.
func NewPipelineQueued(opts ...PipelineQueuedOption) (*PipelineQueued, error) {
	// Defaults
	e := &PipelineQueued{
		PipelineExecutionID: util.NewPipelineExecutionId(),
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

// ForPipelineQueue returns a PipelineQueuedOption that sets the fields of the
// PipelineQueued event from a PipelineQueue command.
func ForPipelineQueue(cmd *PipelineQueue) PipelineQueuedOption {
	return func(e *PipelineQueued) error {
		e.Event = NewFlowEvent(cmd.Event)
		e.Name = cmd.Name
		e.Args = cmd.Args
		if cmd.PipelineExecutionID != "" {
			// Only overwrite the default execution ID if we've been given one to use
			e.PipelineExecutionID = cmd.PipelineExecutionID
		}
		e.ParentStepExecutionID = cmd.ParentStepExecutionID
		e.ParentExecutionID = cmd.ParentExecutionID
		return nil
	}
}
