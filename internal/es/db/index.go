package db

import (
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

// Ristretto backed pipeline datatabase

func GetPipeline(name string) (*pipeline.Pipeline, error) {
	pipelineCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, pcerr.NotFoundWithMessage("pipeline not found: " + name)
	}

	pipeline, ok := pipelineCached.(*pipeline.Pipeline)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid pipeline")
	}
	return pipeline, nil
}

func ListAllPipelines() ([]pipeline.Pipeline, error) {

	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, pcerr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []pipeline.Pipeline
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, *pipeline)
	}

	return pipelines, nil
}
