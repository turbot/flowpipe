package execution

import (
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

func (ex *Execution) PipelineDefinition(pipelineExecutionID string) (*modconfig.Pipeline, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, perr.BadRequestWithMessage("pipeline execution " + pipelineExecutionID + " not found")
	}

	pipeline, err := db.GetPipeline(pe.Name)

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}
