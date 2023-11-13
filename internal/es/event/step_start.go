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
	StepRetry   *modconfig.StepRetry   `json:"step_retry,omitempty"`

	NextStepAction modconfig.NextStepAction `json:"next_step_action,omitempty"`
}

func (e *StepStart) GetEvent() *Event {
	return e.Event
}

func (e *StepStart) HandlerName() string {
	return CommandStepStart
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

func NewStepStartFromStepQueued(e *StepQueued) (*StepStart, error) {

	cmd := &StepStart{
		Event: NewChildEvent(e.Event),
	}
	if e.StepExecutionID == "" {
		return nil, perr.BadRequestWithMessage("missing step execution ID in pipeline step queued event")
	}

	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	} else {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline step queued: %v", e))
	}
	cmd.StepExecutionID = e.StepExecutionID

	cmd.StepName = e.StepName
	cmd.StepInput = e.StepInput
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.NextStepAction = e.NextStepAction

	return cmd, nil
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

func WithStepLoop(stepLoop *modconfig.StepLoop) StepStartOption {
	return func(cmd *StepStart) error {
		cmd.StepLoop = stepLoop
		return nil
	}
}
