package event

import (
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
)

// PipelineQueue commands a pipeline to be queued for execution.
type PipelineQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline details
	Name string          `json:"name"`
	Args modconfig.Input `json:"args" cty:"args"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// If this is a child pipeline then set the parent pipeline execution ID
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
	ParentExecutionID     string `json:"parent_execution_id,omitempty"`
}

func (e *PipelineQueue) GetEvent() *Event {
	return e.Event
}

func (e *PipelineQueue) HandlerName() string {
	return "command.pipeline_queue"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineQueueOption func(*PipelineQueue) error

// NewPipelineQueue creates a new PipelineQueue event.
func NewPipelineQueue(opts ...PipelineQueueOption) (*PipelineQueue, error) {
	// Defaults
	e := &PipelineQueue{
		PipelineExecutionID: util.NewPipelineExecutionID(),
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

// ForPipelineQueue returns a PipelineQueueOption that sets the fields of the
// PipelineQueue event from a PipelineQueue command.
func ForPipelineStepStartedToPipelineQueue(e *StepPipelineStarted) PipelineQueueOption {
	return func(cmd *PipelineQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.PipelineExecutionID = e.ChildPipelineExecutionID

		cmd.ParentStepExecutionID = e.StepExecutionID
		cmd.ParentExecutionID = e.PipelineExecutionID

		cmd.Name = e.ChildPipelineName
		cmd.Args = e.ChildPipelineArgs
		return nil
	}
}
