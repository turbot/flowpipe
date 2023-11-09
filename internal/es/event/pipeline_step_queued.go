package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStepQueued struct {
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

func (e *PipelineStepQueued) GetEvent() *Event {
	return e.Event
}

func (e *PipelineStepQueued) HandlerName() string {
	return "handler.pipeline_step_queued"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepQueuedOption func(*PipelineStepQueued) error

// NewPipelineStepQueued creates a new PipelineStepQueued event.
func NewPipelineStepQueued(opts ...PipelineStepQueuedOption) (*PipelineStepQueued, error) {
	// Defaults
	e := &PipelineStepQueued{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineStepQueue(cmd *PipelineStepQueue) PipelineStepQueuedOption {
	return func(e *PipelineStepQueued) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline start command: %v", cmd))
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
