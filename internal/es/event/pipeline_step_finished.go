package event

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
)

type PipelineStepFinished struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`

	// Output from the primitive, this is the "native" output of the primitive
	Output *modconfig.Output `json:"output,omitempty"`

	// Step output configured from the output block, we need a separate field for this because
	// we don't want the StepOutput accidentally override the native primtive outputs
	StepOutput map[string]interface{} `json:"step_output,omitempty"`

	// loop controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *modconfig.StepLoop    `json:"step_loop,omitempty"`
}

func (e *PipelineStepFinished) GetEvent() *Event {
	return e.Event
}

func (e *PipelineStepFinished) HandlerName() string {
	return "handler.pipeline_step_finished"
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelineStepFinishedOption func(*PipelineStepFinished) error

// NewPipelineStepFinished creates a new PipelineStepFinished event.
func NewPipelineStepFinished(opts ...PipelineStepFinishedOption) (*PipelineStepFinished, error) {
	// Defaults
	e := &PipelineStepFinished{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func ForPipelineStepStartToPipelineStepFinished(cmd *StepStart) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline step start command: %v", e)
		}
		if cmd.StepExecutionID != "" {
			e.StepExecutionID = cmd.StepExecutionID
		} else {
			return fmt.Errorf("missing step execution ID in pipeline step start command: %v", e)
		}
		e.StepForEach = cmd.StepForEach
		return nil
	}
}

func ForPipelineStepFinish(cmd *PipelineStepFinish) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline step finish command: %v", e)
		}
		if cmd.StepExecutionID != "" {
			e.StepExecutionID = cmd.StepExecutionID
		} else {
			return fmt.Errorf("missing step execution ID in pipeline step finish command: %v", e)
		}
		e.Output = cmd.Output
		e.StepForEach = cmd.StepForEach
		return nil
	}
}

func WithStepOutput(output *modconfig.Output, stepOutput map[string]interface{}, stepLoop *modconfig.StepLoop) PipelineStepFinishedOption {
	return func(e *PipelineStepFinished) error {
		e.Output = output
		e.StepOutput = stepOutput
		e.StepLoop = stepLoop
		return nil
	}
}
