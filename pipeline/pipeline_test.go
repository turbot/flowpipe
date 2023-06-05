package pipeline

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/types"
	"gopkg.in/yaml.v2"
)

func TestNewExecution(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	cfg, err := config.NewConfig(ctx)
	assert.Nil(err)

	cfg.LogDir = "./tmp"
	assert.Equal(cfg.LogDir, "./tmp")

	data, err := os.ReadFile("./series_of_for_loop_steps.yaml")
	assert.Nil(err)

	var pipeline *types.Pipeline
	err = yaml.Unmarshal(data, &pipeline)
	assert.Nil(err)
	assert.Equal("series_of_for_loop_steps", pipeline.Name)
	assert.NotNil(pipeline.Steps["http_1"])
	assert.NotNil(pipeline.Steps["sleep_1"])

	assert.Equal("http_request", pipeline.Steps["http_1"].Type)
}

func TestLoadPipelineDir(t *testing.T) {
	assert := assert.New(t)
	pipelines, err := LoadPipelines(context.Background(), "./")
	assert.Nil(err)
	assert.Len(pipelines, 3)

	pipeline := pipelines[0]
	assert.Equal("for_loop_using_http_request_body_json", pipeline.Name)
}
