package event

import (
	"github.com/turbot/flowpipe/internal/resources"
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
	Output              *resources.Output `json:"output,omitempty"`

	// for_each controls
	StepForEach *resources.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *resources.StepLoop    `json:"step_loop,omitempty"`
	StepRetry   *resources.StepRetry   `json:"step_retry,omitempty"`
	StepInput   resources.Input        `json:"step_input,omitempty"`
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
		cmd.Output = &resources.Output{
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
		cmd.Output = &resources.Output{
			Status: "failed",
			Data: map[string]interface{}{
				"output": e.PipelineOutput,
			},
			Errors: e.Errors,
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

func WithStepForEach(stepForEach *resources.StepForEach) StepPipelineFinishOption {
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
