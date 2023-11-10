package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueue struct {
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
	StepRetry   *modconfig.StepRetry   `json:"step_retry,omitempty"`

	DelayMs int `json:"delay_ms,omitempty"` // delay start in milliseconds

	NextStepAction modconfig.NextStepAction `json:"action,omitempty"`
}

func (e *StepQueue) GetEvent() *Event {
	return e.Event
}

func (e *StepQueue) HandlerName() string {
	return CommandStepQueue
}

type StepQueueOption func(*StepQueue) error

// NewStepQueue creates a new StepQueue event.
func NewStepQueue(opts ...StepQueueOption) (*StepQueue, error) {
	// Defaults
	e := &StepQueue{
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

func NewStepQueueFromPipelineStepFinished(e *StepFinished, stepName string) *StepQueue {

	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionID(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	}

	cmd.StepName = stepName
	cmd.StepInput = *e.StepLoop.Input
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.DelayMs = 0
	cmd.NextStepAction = modconfig.NextStepActionStart

	return cmd
}

func NewStepQueueFromStepForEachPlanned(e *StepForEachPlanned, nextStep *modconfig.NextStep) (*StepQueue, error) {
	cmd := &StepQueue{
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

func StepQueueForPipelinePlanned(e *PipelinePlanned) StepQueueOption {
	return func(cmd *StepQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
		}
		return nil
	}
}

func StepQueueWithStep(name string, input modconfig.Input, stepForEach *modconfig.StepForEach, stepLoop *modconfig.StepLoop, delayMs int, nextStepAction modconfig.NextStepAction) StepQueueOption {
	return func(cmd *StepQueue) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.StepForEach = stepForEach
		cmd.StepLoop = stepLoop
		cmd.DelayMs = delayMs
		cmd.NextStepAction = nextStepAction
		return nil
	}
}
