package event

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type PipelineLoaded struct {
	// Event metadata
	Event *Event `json:"event"`
	// Unique identifier for this pipeline execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	// Pipeline definition that was loaded
	Pipeline *flowpipe.Pipeline `json:"pipeline"`
}

func (e *PipelineLoaded) GetEvent() *Event {
	return e.Event
}

func (e *PipelineLoaded) HandlerName() string {
	return HandlerPipelineLoaded
}

// NewPipelineLoaded creates a new PipelineLoaded event.
func NewPipelineLoadedFromPipelineLoad(cmd *PipelineLoad, pipeline *flowpipe.Pipeline) *PipelineLoaded {
	// Defaults
	e := &PipelineLoaded{
		Event:               NewFlowEvent(cmd.Event),
		PipelineExecutionID: cmd.PipelineExecutionID,
		Pipeline:            pipeline,
	}
	return e
}
