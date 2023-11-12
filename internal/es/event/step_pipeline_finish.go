package event

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

// There's only one use case for this, which is to handle the "Pipeline Step" finish **command**.
//
// Do not confuse/conflate this with the step_finished **event** which is raised when a step has finished.
type StepPipelineFinish struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string            `json:"pipeline_execution_id"`
	StepExecutionID     string            `json:"step_execution_id"`
	Output              *modconfig.Output `json:"output,omitempty"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
}

func (e *StepPipelineFinish) GetEvent() *Event {
	return e.Event
}

func (e *StepPipelineFinish) HandlerName() string {
	return CommandStepPipelineFinish
}

type StepPipelineFinishOption func(*StepPipelineFinish) error

// NewStepPipelineFinish creates a new StepPipelineFinish event.
func NewStepPipelineFinish(opts ...StepPipelineFinishOption) (*StepPipelineFinish, error) {
	// Defaults
	e := &StepPipelineFinish{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineFinished(e *PipelineFinished) StepPipelineFinishOption {
	return func(cmd *StepPipelineFinish) error {
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

func ForPipelineFailed(e *PipelineFailed) StepPipelineFinishOption {
	return func(cmd *StepPipelineFinish) error {
		cmd.Event = NewChildEvent(e.Event)
		cmd.Output = &modconfig.Output{
			Status: "failed",
			Data: map[string]interface{}{
				"output": e.PipelineOutput,
			},
			Errors: []modconfig.StepError{*e.Error},
		}

		return nil
	}
}

func WithPipelineExecutionID(id string) StepPipelineFinishOption {
	return func(cmd *StepPipelineFinish) error {
		cmd.PipelineExecutionID = id
		return nil
	}
}

func WithStepForEach(stepForEach *modconfig.StepForEach) StepPipelineFinishOption {
	return func(cmd *StepPipelineFinish) error {
		cmd.StepForEach = stepForEach
		return nil
	}
}

func WithStepExecutionID(id string) StepPipelineFinishOption {
	return func(cmd *StepPipelineFinish) error {
		cmd.StepExecutionID = id
		return nil
	}
}
