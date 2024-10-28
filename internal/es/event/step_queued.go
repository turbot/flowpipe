package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueued struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`

	StepExecutionID string         `json:"step_execution_id"`
	StepName        string         `json:"step_name"`
	StepType        string         `json:"step_type"`
	StepInput       flowpipe.Input `json:"input"`

	// for_each controls
	StepForEach    *flowpipe.StepForEach   `json:"step_for_each,omitempty"`
	StepLoop       *flowpipe.StepLoop      `json:"step_loop,omitempty"`
	StepRetry      *flowpipe.StepRetry     `json:"step_retry,omitempty"`
	NextStepAction flowpipe.NextStepAction `json:"next_step_action,omitempty"`
}

func (e *StepQueued) GetEvent() *Event {
	return e.Event
}

func (e *StepQueued) HandlerName() string {
	return HandlerStepQueued
}

type StepQueuedOption func(*StepQueued) error

// NewStepQueued creates a new StepQueued event.
func NewStepQueued(opts ...StepQueuedOption) (*StepQueued, error) {
	// Defaults
	e := &StepQueued{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForStepQueue(cmd *StepQueue) StepQueuedOption {
	return func(e *StepQueued) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline step queued: %v", cmd))
		}
		e.StepExecutionID = cmd.StepExecutionID
		e.StepName = cmd.StepName
		e.StepInput = cmd.StepInput
		e.StepForEach = cmd.StepForEach
		e.StepLoop = cmd.StepLoop
		e.StepRetry = cmd.StepRetry
		e.NextStepAction = cmd.NextStepAction
		return nil
	}
}
