package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestStepErrorConfig(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/http.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["bad_http"] == nil {
		assert.Fail("bad_http pipeline not found")
		return
	}

}
