package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string         `json:"pipeline_execution_id"`
	StepExecutionID     string         `json:"step_execution_id"`
	StepName            string         `json:"step_name"`
	StepInput           flowpipe.Input `json:"input"`

	// for_each controls
	StepForEach    *flowpipe.StepForEach `json:"step_for_each,omitempty"`
	StepLoop       *flowpipe.StepLoop    `json:"step_loop,omitempty"`
	StepRetry      *flowpipe.StepRetry   `json:"step_retry,omitempty"`
	MaxConcurrency *int                   `json:"max_concurrency,omitempty"`

	NextStepAction flowpipe.NextStepAction `json:"action,omitempty"`
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
		StepExecutionID: util.NewStepExecutionId(),
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

func NewStepQueueFromPipelineStepFinishedForLoop(e *StepFinished, stepName string) *StepQueue {

	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionId(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	}

	extendedInput := util.ExtendInputs(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, stepName, *e.StepLoop.Input)

	cmd.StepName = stepName
	cmd.StepInput = extendedInput
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.NextStepAction = flowpipe.NextStepActionStart

	return cmd
}

func NewStepQueueFromPipelineStepFinishedForRetry(e *StepFinished, stepName string) *StepQueue {

	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionId(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	}

	cmd.StepName = stepName
	cmd.StepInput = *e.StepRetry.Input
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.NextStepAction = flowpipe.NextStepActionStart

	return cmd
}

func NewStepQueueFromStepForEachPlanned(e *StepForEachPlanned, nextStep *flowpipe.NextStep) (*StepQueue, error) {
	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionId(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	} else {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
	}

	extendedInput := util.ExtendInputs(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, e.StepName, nextStep.Input)
	cmd.StepName = e.StepName
	cmd.StepInput = extendedInput
	cmd.StepForEach = nextStep.StepForEach
	cmd.StepLoop = nil
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

func StepQueueWithStep(name string, input flowpipe.Input, stepForEach *flowpipe.StepForEach, stepLoop *flowpipe.StepLoop, nextStepAction flowpipe.NextStepAction) StepQueueOption {
	return func(cmd *StepQueue) error {
		extendedInput := util.ExtendInputs(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID, name, input)
		cmd.StepName = name
		cmd.StepInput = extendedInput
		cmd.StepForEach = stepForEach
		cmd.StepLoop = stepLoop
		cmd.NextStepAction = nextStepAction
		return nil
	}
}
