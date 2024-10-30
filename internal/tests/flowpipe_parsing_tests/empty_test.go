package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
)

func TestEmptySlice(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/empty.fp")
	assert.Nil(err, "error found")
	assert.NotNil(pipelines, "pipelines not found")

	pipeline := pipelines["local.pipeline.empty_slice"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	input, err := pipeline.Steps[0].GetInputs(nil)
	assert.Nil(err)

	if input == nil {
		assert.Fail("input not found")
		return
	}

	value := input["value"]
	if value == nil {
		assert.Fail("value not found")
		return
	}

	valueSlice, ok := value.([]interface{})
	if !ok {
		assert.Fail("value is not a slice")
		return
	}

	if valueSlice == nil {
		assert.Fail("value slice is nil")
		return
	}

	assert.Equal(0, len(valueSlice))
}
