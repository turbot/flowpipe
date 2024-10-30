package event

import (
	"fmt"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
)

// StepFinished event is when a step (any step step has completed). This is an event that will be handled
// by an Event Handler (StepFinishedHandler)
//
// Do not confuse this with Pipeline Step Finish **command** which is raised when a child pipeline has finished
type StepFinished struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`

	// Output from the primitive, this is the "native" output of the primitive
	Output *resources.Output `json:"output,omitempty"`

	// Step output configured from the output block, we need a separate field for this because
	// we don't want the StepOutput accidentally override the native primtive outputs
	StepOutput map[string]interface{} `json:"step_output,omitempty"`

	// loop controls
	StepForEach *resources.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *resources.StepLoop    `json:"step_loop,omitempty"`
	StepRetry   *resources.StepRetry   `json:"step_retry,omitempty"`
}

func (e *StepFinished) GetEvent() *Event {
	return e.Event
}

func (e *StepFinished) HandlerName() string {
	return HandlerStepFinished
}

type StepFinishedOption func(*StepFinished) error

// NewStepFinished creates a new StepFinished event.
func NewStepFinished(opts ...StepFinishedOption) (*StepFinished, error) {
	// Defaults
	e := &StepFinished{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func NewStepFinishedFromStepStart(cmd *StepStart, output *resources.Output, stepOutput map[string]interface{}, stepLoop *resources.StepLoop) (*StepFinished, error) {
	e := StepFinished{
		Event: NewFlowEvent(cmd.Event),
	}
	if cmd.PipelineExecutionID != "" {
		e.PipelineExecutionID = cmd.PipelineExecutionID
	} else {
		return nil, perr.BadRequestWithMessage("missing pipeline execution ID in pipeline step start command")
	}
	if cmd.StepExecutionID != "" {
		e.StepExecutionID = cmd.StepExecutionID
	} else {
		return nil, perr.BadRequestWithMessage("missing step execution ID in pipeline step start command")
	}
	e.StepForEach = cmd.StepForEach

	e.Output = output
	e.StepOutput = stepOutput
	e.StepLoop = stepLoop

	return &e, nil
}

func ForPipelineStepFinish(cmd *StepPipelineFinish) StepFinishedOption {
	return func(e *StepFinished) error {
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
