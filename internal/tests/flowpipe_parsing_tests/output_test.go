package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
)

func TestOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/output.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.with_output"] == nil {
		assert.Fail("with_output pipeline not found")
		return
	}

	if len(pipelines["local.pipeline.with_output"].OutputConfig) != 2 {
		assert.Fail("with_output pipeline has no outputs")
		return
	}

	outputs := pipelines["local.pipeline.with_output"].OutputConfig
	assert.Equal("one", outputs[0].Name)
	assert.Equal("two", outputs[1].Name)
}
