package event

import (
	"fmt"
)

type PipelinePlan struct {
	// Event metadata
	Event *Event `json:"event"`
	// Pipeline execution details
	PipelineExecutionID string `json:"pipeline_execution_id"`
}

func (e *PipelinePlan) GetEvent() *Event {
	return e.Event
}

func (e *PipelinePlan) HandlerName() string {
	return CommandPipelinePlan
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelinePlanOption func(*PipelinePlan) error

// NewPipelinePlan creates a new PipelinePlan event.
func NewPipelinePlan(opts ...PipelinePlanOption) (*PipelinePlan, error) {
	// Defaults
	e := &PipelinePlan{}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func NewPipelinePlanFromStepForEachPlanned(e *StepForEachPlanned) *PipelinePlan {
	cmd := &PipelinePlan{
		Event:               NewChildEvent(e.Event),
		PipelineExecutionID: e.PipelineExecutionID,
	}
	return cmd
}

func ForPipelineStarted(e *PipelineStarted) PipelinePlanOption {
	return func(cmd *PipelinePlan) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline loaded event: %v", e)
		}
		return nil
	}
}

func ForPipelineResumed(e *PipelineResumed) PipelinePlanOption {
	return func(cmd *PipelinePlan) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline loaded event: %v", e)
		}
		return nil
	}
}

func ForPipelineStepFinished(e *StepFinished) PipelinePlanOption {
	return func(cmd *PipelinePlan) error {
		cmd.Event = NewFlowEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return fmt.Errorf("missing pipeline execution ID in pipeline step finished event: %v", e)
		}
		return nil
	}
}

func PipelinePlanFromPipelinePlanned(e *PipelinePlanned) *PipelinePlan {
	cmd := &PipelinePlan{
		Event:               NewFlowEvent(e.Event),
		PipelineExecutionID: e.PipelineExecutionID,
	}
	return cmd
}

func ForChildPipelineFinished(e *PipelineFinished, parentPipelineExecutionID string) PipelinePlanOption {
	return func(cmd *PipelinePlan) error {
		cmd.Event = NewFlowEvent(e.Event)
		cmd.PipelineExecutionID = parentPipelineExecutionID
		return nil
	}
}
