package event

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

type PipelineStepFinish struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string            `json:"pipeline_execution_id"`
	StepExecutionID     string            `json:"step_execution_id"`
	Output              *modconfig.Output `json:"output,omitempty"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
}

func (e *PipelineStepFinish) GetEvent() *Event {
	return e.Event
}

func (e *PipelineStepFinish) HandlerName() string {
	return "command.pipeline_step_finish"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepFinishOption func(*PipelineStepFinish) error

// NewPipelineStepFinish creates a new PipelineStepFinish event.
func NewPipelineStepFinish(opts ...PipelineStepFinishOption) (*PipelineStepFinish, error) {
	// Defaults
	e := &PipelineStepFinish{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineFinished(e *PipelineFinished) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.Output = &modconfig.Output{
			Status: "", // output is only relevant for step
			Data: map[string]interface{}{
				"output": e.PipelineOutput,
			},
		}

		return nil
	}
}

func ForPipelineFailed(e *PipelineFailed) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.Output = &modconfig.Output{
			Status: "",
			Data: map[string]interface{}{
				"output": e.PipelineOutput,
			},
			Errors: []modconfig.StepError{*e.Error},
		}

		return nil
	}
}

func WithPipelineExecutionID(id string) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.PipelineExecutionID = id
		return nil
	}
}

func WithStepForEach(stepForEach *modconfig.StepForEach) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.StepForEach = stepForEach
		return nil
	}
}

func WithStepExecutionID(id string) PipelineStepFinishOption {
	return func(cmd *PipelineStepFinish) error {
		cmd.StepExecutionID = id
		return nil
	}
}
