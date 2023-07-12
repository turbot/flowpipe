package db

import (
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/types"
)

// Ristretto backed pipeline datatabase

func GetPipeline(name string) (*types.PipelineHcl, error) {
	pipelineCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, fperr.NotFoundWithMessage("pipeline " + name + " not found")
	}

	pipeline, ok := pipelineCached.(*types.PipelineHcl)
	if !ok {
		return nil, fperr.InternalWithMessage("invalid pipeline")
	}

	// When we start the API server we load the entire pipelines given in the start up command line and store it in memory.
	// And, the GET /pipeline operation just dump the content in the response.
	// For now just return the name and description of the pipeline, not even the steps and we'll start adding what else should we return in the API.
	getPipelineOutput := &types.PipelineHcl{
		Name:        pipeline.Name,
		Description: pipeline.Description,
	}
	return getPipelineOutput, nil
}

func ListAllPipelines() ([]types.PipelineHcl, error) {

	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, fperr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, fperr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []types.PipelineHcl
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, *pipeline)
	}

	return pipelines, nil
}
