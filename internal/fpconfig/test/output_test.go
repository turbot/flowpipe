package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fpconfig"
)

func TestOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := fpconfig.LoadPipelines(context.TODO(), "./test_pipelines/output.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["with_output"] == nil {
		assert.Fail("with_output pipeline not found")
		return
	}

	if len(pipelines["with_output"].Outputs) != 2 {
		assert.Fail("with_output pipeline has no outputs")
		return
	}

	outputs := pipelines["with_output"].Outputs
	assert.Equal("one", outputs[0].Name)
	assert.Equal("two", outputs[1].Name)
}
