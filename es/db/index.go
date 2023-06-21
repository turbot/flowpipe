package db

import (
	"github.com/turbot/flowpipe/cache"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/types"
)

// Ristretto backed pipeline datatabase

func GetPipeline(name string) (*types.Pipeline, error) {
	pipelineCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, fperr.NotFoundWithMessage("pipeline " + name + " not found")
	}

	pipeline, ok := pipelineCached.(*types.Pipeline)
	if !ok {
		return nil, fperr.InternalWithMessage("invalid pipeline")
	}
	return pipeline, nil
}

func ListAllPipelines() ([]types.Pipeline, error) {

	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, fperr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, fperr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []types.Pipeline
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, *pipeline)
	}

	return pipelines, nil
}
