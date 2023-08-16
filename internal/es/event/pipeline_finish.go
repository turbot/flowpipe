package event

import (
	"fmt"

	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

type PipelineFinish struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string           `json:"pipeline_execution_id"`
	Output              *pipeline.Output `json:"output,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineFinishOption func(*PipelineFinish) error

// NewPipelineFinish creates a new PipelineFinish event.
func NewPipelineFinish(opts ...PipelineFinishOption) (*PipelineFinish, error) {
	// Defaults
	e := &PipelineFinish{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelinePlannedToPipelineFinish(e *PipelinePlanned) PipelineFinishOption {
	return func(cmd *PipelineFinish) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		return nil
	}
}

func WithPipelineOutput(output *pipeline.Output) PipelineFinishOption {
	return func(e *PipelineFinish) error {
		e.Output = output
		return nil
	}
}
