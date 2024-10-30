package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/retry.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.retry_simple"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	assert.NotNil(pipeline.Steps[0].GetRetryConfig(nil, false))
	retryConfig, diags := pipeline.Steps[0].GetRetryConfig(nil, false)
	assert.Equal(0, len(diags))
	assert.Equal(int64(2), *retryConfig.MaxAttempts)
	assert.Equal("exponential", *retryConfig.Strategy)

	pipeline = pipelines["local.pipeline.retry_with_if"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")

	pipeline = pipelines["local.pipeline.retry_default"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	assert.NotNil(pipeline.Steps, "steps not found")
	retryConfig, diags = pipeline.Steps[0].GetRetryConfig(nil, false)
	assert.Equal(0, len(diags))
	assert.NotNil(retryConfig)
	assert.Nil(retryConfig.MaxAttempts)
	assert.Nil(retryConfig.Strategy)
	assert.Nil(retryConfig.MinInterval)
	assert.Nil(retryConfig.MaxInterval)

}

func TestRetryWithBackoff(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/retry_with_backoff.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.retry_with_default_backoff"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	retryConfig, diags := pipeline.Steps[0].GetRetryConfig(nil, false)

	if len(diags) > 0 {
		assert.Fail("diags found", diags)
		return
	}

	// constant backoff, always the min interval: 1000ms
	assert.Equal(int64(0), retryConfig.CalculateBackoff(1).Milliseconds())
	assert.Equal(int64(1000), retryConfig.CalculateBackoff(2).Milliseconds())
	assert.Equal(int64(1000), retryConfig.CalculateBackoff(3).Milliseconds())

	pipeline = pipelines["local.pipeline.retry_with_linear_backoff"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	retryConfig, diags = pipeline.Steps[0].GetRetryConfig(nil, false)

	if len(diags) > 0 {
		assert.Fail("diags found", diags)
		return
	}

	// linear backoff with interval of 500ms
	assert.Equal(int64(0), retryConfig.CalculateBackoff(1).Milliseconds())
	assert.Equal(int64(500), retryConfig.CalculateBackoff(2).Milliseconds())
	assert.Equal(int64(1000), retryConfig.CalculateBackoff(3).Milliseconds())
	assert.Equal(int64(1500), retryConfig.CalculateBackoff(4).Milliseconds())
	assert.Equal(int64(2000), retryConfig.CalculateBackoff(5).Milliseconds())

	// max interval is 4000
	assert.Equal(int64(4000), retryConfig.CalculateBackoff(100).Milliseconds())

	pipeline = pipelines["local.pipeline.retry_with_exponential_backoff"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	retryConfig, diags = pipeline.Steps[0].GetRetryConfig(nil, false)

	if len(diags) > 0 {
		assert.Fail("diags found", diags)
		return
	}

	// exponential backoff with interval of 500ms
	assert.Equal(int64(0), retryConfig.CalculateBackoff(1).Milliseconds())
	assert.Equal(int64(500), retryConfig.CalculateBackoff(2).Milliseconds())
	assert.Equal(int64(1000), retryConfig.CalculateBackoff(3).Milliseconds())
	assert.Equal(int64(2000), retryConfig.CalculateBackoff(4).Milliseconds())
	assert.Equal(int64(4000), retryConfig.CalculateBackoff(5).Milliseconds())
	assert.Equal(int64(8000), retryConfig.CalculateBackoff(6).Milliseconds())
	assert.Equal(int64(32000), retryConfig.CalculateBackoff(8).Milliseconds())

	// max interval is 50000
	assert.Equal(int64(50000), retryConfig.CalculateBackoff(10).Milliseconds())
}
