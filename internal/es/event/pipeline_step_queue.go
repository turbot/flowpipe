package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

type PipelineStepQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string         `json:"pipeline_execution_id"`
	StepExecutionID     string         `json:"step_execution_id"`
	StepName            string         `json:"step_name"`
	StepInput           pipeline.Input `json:"input"`

	// for_each controls
	StepForEach *pipeline.StepForEach `json:"step_for_each,omitempty"`

	DelayMs int `json:"delay_ms,omitempty"` // delay start in milliseconds

	NextStepAction pipeline.NextStepAction `json:"action,omitempty"`
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

func PipelineStepQueueWithStep(name string, input pipeline.Input, stepForEach *pipeline.StepForEach, delayMs int, nextStepAction pipeline.NextStepAction) PipelineStepQueueOption {
	return func(cmd *PipelineStepQueue) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.StepForEach = stepForEach
		cmd.DelayMs = delayMs
		cmd.NextStepAction = nextStepAction
		return nil
	}
}
