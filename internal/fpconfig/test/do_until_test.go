package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestDoUntil(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/do_until.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.do_until"] == nil {
		assert.Fail("do_until pipeline not found")
		return
	}
}
