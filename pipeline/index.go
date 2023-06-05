package pipeline

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/flowpipe/cache"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/types"
	"gopkg.in/yaml.v2"
)

func LoadPipelines(ctx context.Context, directory string) ([]types.Pipeline, error) {
	var data []types.Pipeline

	fplog.Logger(ctx).Debug("Loading pipelines", "directory", directory)

	// Read directory contents
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	inMemoryCache := cache.GetCache()
	// Iterate over files
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext != ".yaml" && ext != ".yml" {
				continue
			}

			// Get the file path
			filePath := filepath.Join(directory, file.Name())

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
			var pipeline types.Pipeline
			err = yaml.Unmarshal(fileBytes, &pipeline)
			if err != nil {
				return nil, err
			}

			fplog.Logger(ctx).Debug("Loaded pipeline", "name", pipeline.Name, "file", filePath)

			// Append to data slice
			data = append(data, pipeline)

			inMemoryCache.SetWithTTL(pipeline.Name, pipeline, 24*7*52*99*time.Hour)
		}
	}

	return data, nil
}
