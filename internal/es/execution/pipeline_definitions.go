package execution

import (
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
)

func (ex *Execution) PipelineDefinition(pipelineExecutionID string) (*resources.Pipeline, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, perr.BadRequestWithMessage("pipeline execution " + pipelineExecutionID + " not found")
	}

	pipeline, err := db.GetPipelineWithModFullVersion(pe.ModFullVersion, pe.Name)

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}
