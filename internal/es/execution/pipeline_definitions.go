package execution

import (
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

func (ex *Execution) PipelineDefinition(pipelineExecutionID string) (*pipeline.Pipeline, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, pcerr.BadRequestWithMessage("pipeline execution " + pipelineExecutionID + " not found")
	}

	pipeline, err := db.GetPipeline(pe.Name)

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}
