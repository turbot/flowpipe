package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/schema"
)

func TestStepErrorConfig(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/error.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.bad_http"] == nil {
		assert.Fail("bad_http pipeline not found")
		return
	}
}

func TestStepErrorConfigWithIf(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/error.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.bad_http_ignored_with_if"] == nil {
		assert.Fail("bad_http pipeline not found")
		return
	}

	pipeline := pipelines["local.pipeline.bad_http_ignored_with_if"]

	step := pipeline.GetStep("http.my_step_1")
	errConfig, diags := step.GetErrorConfig(nil, false)
	assert.False(diags.HasErrors(), "diags has errors")

	assert.NotNil(errConfig.UnresolvedAttributes[schema.AttributeTypeIf], "if attribute not found")
}

func TestStepErrorConfigRetries(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/error.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.bad_http_retries"] == nil {
		assert.Fail("bad_http_retries pipeline not found")
		return
	}

	step := pipelines["local.pipeline.bad_http_retries"].GetStep("http.my_step_1")

	if step == nil {
		assert.Fail("step not found")
		return
	}

	errorConfig, diags := step.GetErrorConfig(nil, true)
	if diags.HasErrors() {
		assert.Fail("diags has errors")
		return
	}

	if errorConfig == nil {
		assert.Fail("error config not found")
		return
	}
}
