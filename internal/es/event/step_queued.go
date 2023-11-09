package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueued struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`

	StepExecutionID string          `json:"step_execution_id"`
	StepName        string          `json:"step_name"`
	StepInput       modconfig.Input `json:"input"`

	// for_each controls
	StepForEach    *modconfig.StepForEach   `json:"step_for_each,omitempty"`
	StepLoop       *modconfig.StepLoop      `json:"step_loop,omitempty"`
	NextStepAction modconfig.NextStepAction `json:"next_step_action,omitempty"`

	DelayMs int `json:"delay_ms,omitempty"` // delay start in milliseconds
}

func (e *StepQueued) GetEvent() *Event {
	return e.Event
}

func (e *StepQueued) HandlerName() string {
	return "handler.step_queued"
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
		e.NextStepAction = cmd.NextStepAction
		e.DelayMs = cmd.DelayMs
		return nil
	}
}
