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

	pipeline, ok := pipelineCached.(types.Pipeline)
	if !ok {
		return nil, fperr.InternalWithMessage("invalid pipeline")
	}
	return &pipeline, nil
}
