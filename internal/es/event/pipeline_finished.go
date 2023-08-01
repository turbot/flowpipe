package event

import (
	"fmt"
)

type PipelineFinished struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`

	PipelineOutput map[string]interface{} `json:"pipeline_output"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineFinishedOption func(*PipelineFinished) error

// NewPipelineFinished creates a new PipelineFinished event.
func NewPipelineFinished(opts ...PipelineFinishedOption) (*PipelineFinished, error) {
	// Defaults
	e := &PipelineFinished{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineFinish returns a PipelineFinishedOption that sets the fields of the
// PipelineFinished event from a PipelineFinish command.
func ForPipelineFinish(cmd *PipelineFinish, pipelineOutput map[string]interface{}) PipelineFinishedOption {
	return func(e *PipelineFinished) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline start command: %v", cmd)
		}
		e.PipelineOutput = pipelineOutput
		return nil
	}
}
