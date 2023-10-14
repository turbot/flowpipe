package event

import (
	"fmt"

	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type PipelinePlanned struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// The planner outputs a list of the next steps to be executed in the types.
	NextSteps []modconfig.NextStep `json:"next_steps"`
}

// ExecutionOption is a function that modifies an Execution instance.
type PipelinePlannedOption func(*PipelinePlanned) error

// NewPipelinePlanned creates a new PipelinePlanned event.
func NewPipelinePlanned(opts ...PipelinePlannedOption) (*PipelinePlanned, error) {
	// Defaults
	e := &PipelinePlanned{
		NextSteps: []modconfig.NextStep{},
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

// ForPipelinePlan returns a PipelinePlannedOption that sets the fields of the
// PipelinePlanned event from a PipelinePlan command.
func ForPipelinePlan(cmd *PipelinePlan) PipelinePlannedOption {
	return func(e *PipelinePlanned) error {
		e.Event = NewFlowEvent(cmd.Event)
		if cmd.PipelineExecutionID != "" {
			e.PipelineExecutionID = cmd.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline plan command: %v", cmd))
		}
		return nil
	}
}
