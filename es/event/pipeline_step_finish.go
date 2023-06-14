package event

import "github.com/turbot/flowpipe/types"

type PipelineStepFinish struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string            `json:"pipeline_execution_id"`
	StepExecutionID     string            `json:"step_execution_id"`
	Output              *types.StepOutput `json:"output,omitempty"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepFinishOption func(*PipelineStepFinish) error

// NewPipelineStepFinish creates a new PipelineStepFinish event.
func NewPipelineStepFinish(opts ...PipelineStepFinishOption) (*PipelineStepFinish, error) {
	// Defaults
	e := &PipelineStepFinish{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineFinished(e *PipelineFinished) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.Output = e.Output
		return nil
	}
}

func WithPipelineExecutionID(id string) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.PipelineExecutionID = id
		return nil
	}
}

func WithStepExecutionID(id string) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.StepExecutionID = id
		return nil
	}
}
