package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/types"
)

type PipelineLoaded struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Pipeline definition that was loaded
	Pipeline *types.Pipeline `json:"pipeline"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineLoadedOption func(*PipelineLoaded) error

// NewPipelineLoaded creates a new PipelineLoaded event.
func NewPipelineLoaded(opts ...PipelineLoadedOption) (*PipelineLoaded, error) {
	// Defaults
	e := &PipelineLoaded{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

// ForPipelineLoad returns a PipelineLoadedOption that sets the fields of the
// PipelineLoaded event from a PipelineLoad command.
func ForPipelineLoad(cmd *PipelineLoad) PipelineLoadedOption {
	return func(e *PipelineLoaded) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline load command: %v", cmd)
		}
		return nil
	}
}

// WithPipeline sets the Pipeline of the PipelineLoaded event.
func WithPipelineDefinition(pipeline *types.Pipeline) PipelineLoadedOption {
	return func(e *PipelineLoaded) error {
		e.Pipeline = pipeline
		return nil
	}
}
