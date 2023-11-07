package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStepQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string          `json:"pipeline_execution_id"`
	StepExecutionID     string          `json:"step_execution_id"`
	StepName            string          `json:"step_name"`
	StepInput           modconfig.Input `json:"input"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *modconfig.StepLoop    `json:"step_loop,omitempty"`

	DelayMs int `json:"delay_ms,omitempty"` // delay start in milliseconds

	NextStepAction modconfig.NextStepAction `json:"action,omitempty"`
}

func (e *PipelineStepQueue) GetEvent() *Event {
	return e.Event
}

func (e *PipelineStepQueue) HandlerName() string {
	return "command.pipeline_step_queue"
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

func NewPipelineStepQueueFromStepForEachPlanned(e *StepForEachPlanned, nextStep *modconfig.NextStep) (*PipelineStepQueue, error) {
	cmd := &PipelineStepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionID(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	} else {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
	}

	cmd.StepName = e.StepName
	cmd.StepInput = nextStep.Input
	cmd.StepForEach = nextStep.StepForEach
	cmd.StepLoop = nil
	cmd.DelayMs = 0
	cmd.NextStepAction = nextStep.Action

	return cmd, nil
}

func PipelineStepQueueForPipelinePlanned(e *PipelinePlanned) PipelineStepQueueOption {
	return func(cmd *PipelineStepQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
		}
		return nil
	}
}

func PipelineStepQueueWithStep(name string, input modconfig.Input, stepForEach *modconfig.StepForEach, stepLoop *modconfig.StepLoop, delayMs int, nextStepAction modconfig.NextStepAction) PipelineStepQueueOption {
	return func(cmd *PipelineStepQueue) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.StepForEach = stepForEach
		cmd.StepLoop = stepLoop
		cmd.DelayMs = delayMs
		cmd.NextStepAction = nextStepAction
		return nil
	}
}
