package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepStart struct {
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

	DelayMs        int                      `json:"delay_ms,omitempty"` // delay start in milliseconds
	NextStepAction modconfig.NextStepAction `json:"next_step_action,omitempty"`
}

func (e *StepStart) GetEvent() *Event {
	return e.Event
}

func (e *StepStart) HandlerName() string {
	return "command.step_start"
}

type StepStartOption func(*StepStart) error

// NewStepStart creates a new StepStart command.
func NewStepStart(opts ...StepStartOption) (*StepStart, error) {
	// Defaults
	e := &StepStart{
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

func ForPipelinePlanned(e *PipelinePlanned) StepStartOption {
	return func(cmd *StepStart) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		return nil
	}
}

func ForPipelineStepQueued(e *StepQueued) StepStartOption {
	return func(cmd *StepStart) error {

		if e.StepExecutionID == "" {
			return perr.BadRequestWithMessage("missing step execution ID in pipeline step queued event")
		}

		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline planned event: %v", e)
		}
		cmd.StepExecutionID = e.StepExecutionID
		return nil
	}
}

func WithStep(name string, input modconfig.Input, stepForEach *modconfig.StepForEach, stepLoop *modconfig.StepLoop, nextStepAction modconfig.NextStepAction) StepStartOption {
	return func(cmd *StepStart) error {
		cmd.StepName = name
		cmd.StepInput = input
		cmd.StepForEach = stepForEach
		cmd.StepLoop = stepLoop
		cmd.NextStepAction = nextStepAction
		return nil
	}
}

func WithStepLoop(stepLoop *modconfig.StepLoop) StepStartOption {
	return func(cmd *StepStart) error {
		cmd.StepLoop = stepLoop
		return nil
	}
}
