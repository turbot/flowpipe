package pipeline_legacy

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types_legacy"
	"gopkg.in/yaml.v2"
)

func LoadPipelines(ctx context.Context, directory string) ([]types_legacy.Pipeline, error) {
	var data []types_legacy.Pipeline

	logger := fplog.Logger(ctx)
	logger.Debug("Loading pipelines", "directory", directory)

	// Read directory contents
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	inMemoryCache := cache.GetCache()
	// Iterate over files

	var pipelineNames []string
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			// Get the file path
			filePath := filepath.Join(directory, file.Name())

			var pipeline *types_legacy.Pipeline
			if ext == ".yaml" || ext == ".yml" {
				logger.Info("Loading pipeline", "file", filePath)
				pipeline, err = loadPipelineYaml(filePath)
				if err != nil {
					return nil, err
				}

			} else if ext == ".fp" {
				continue
			} else {
				logger.Warn("Unknown file extension", "file", file.Name())
				continue
			}

			pipelineNames = append(pipelineNames, pipeline.Name)
			logger.Info("Loaded pipeline", "name", pipeline.Name)

			// Append to data slice
			data = append(data, *pipeline)

			// Set in cache
			// TODO: how do we want to do this?
			inMemoryCache.SetWithTTL(pipeline.Name, pipeline, 24*7*52*99*time.Hour)
		}
	}

	inMemoryCache.SetWithTTL("#pipeline.names", pipelineNames, 24*7*52*99*time.Hour)
	return data, nil
}

func loadPipelineYaml(filePath string) (*types_legacy.Pipeline, error) {
	// Open the file
	fileData, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileData.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(fileData)
	if err != nil {
		return nil, err
	}

	// Parse YAML into struct
	var pipeline types_legacy.Pipeline
	err = yaml.Unmarshal(fileBytes, &pipeline)
	if err != nil {
		return nil, err
	}

	return &pipeline, nil
}
