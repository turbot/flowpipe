package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
)

type PipelineStepQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string            `json:"pipeline_execution_id"`
	StepExecutionID     string            `json:"step_execution_id"`
	StepName            string            `json:"step_name"`
	StepInput           types.Input       `json:"input"`
	Index               *int              `json:"index,omitempty"`
	ForEachOutput       *types.StepOutput `json:"for_each_output,omitempty"`
	DelayMs             int               `json:"delay_ms,omitempty"` // delay start in milliseconds
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepQueueOption func(*PipelineStepQueue) error

// NewPipelineStepQueue creates a new PipelineStepQueue event.
func NewPipelineStepQueue(opts ...PipelineStepQueueOption) (*PipelineStepQueue, error) {
	// Defaults
	e := &PipelineStepQueue{
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

func PipelineStepQueueForPipelinePlanned(e *PipelinePlanned) PipelineStepQueueOption {
	return func(cmd *PipelineStepQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		return nil
	}
}

func PipelineStepQueueWithStep(name string, input types.Input, index *int, forEachOutput *types.StepOutput, delayMs int) PipelineStepQueueOption {
	return func(cmd *PipelineStepQueue) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.Index = index
		cmd.ForEachOutput = forEachOutput
		cmd.DelayMs = delayMs
		return nil
	}
}
