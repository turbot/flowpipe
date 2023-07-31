package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestJsonSimple(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/json.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["json"] == nil {
		assert.Fail("json pipeline not found")
		return
	}

	step := pipelines["json"].GetStep("echo.json")
	if step == nil {
		assert.Fail("echo.json step not found")
		return
	}
}

func TestJsonFor(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/json.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["json_for"] == nil {
		assert.Fail("json_for pipeline not found")
		return
	}

	step := pipelines["json_for"].GetStep("echo.json")
	if step == nil {
		assert.Fail("echo.json step not found")
		return
	}
}
