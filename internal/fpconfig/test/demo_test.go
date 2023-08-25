package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestDemoPipeline(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	pipelines, _, err := pipeparser.LoadPipelines(ctx, "./test_pipelines/complex_one.fp")
	assert.Nil(err, "error found")
	assert.NotNil(pipelines)
	assert.NotNil(pipelines["local.pipeline.complex_one"])

}
