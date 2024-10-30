package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonSimple(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/json.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.json"] == nil {
		assert.Fail("json pipeline not found")
		return
	}

	step := pipelines["local.pipeline.json"].GetStep("transform.json")
	if step == nil {
		assert.Fail("transform.json step not found")
		return
	}
}

func TestJsonFor(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/json.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.json_for"] == nil {
		assert.Fail("json_for pipeline not found")
		return
	}

	step := pipelines["local.pipeline.json_for"].GetStep("transform.json")
	if step == nil {
		assert.Fail("transform.json step not found")
		return
	}
}
