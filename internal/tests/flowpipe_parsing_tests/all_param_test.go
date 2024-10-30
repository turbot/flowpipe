package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
)

func TestAllParam(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/all_param.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.all_param"]
	if pipeline == nil {
		assert.Fail("Pipeline not found")
		return
	}

	// all steps must have unresolved attributes
	for _, step := range pipeline.Steps {
		// except echo bazz
		if step.GetName() == "echo_baz" {
			assert.Nil(step.GetUnresolvedAttributes()["value"])
		} else {
			assert.NotNil(step.GetUnresolvedAttributes()["value"])
		}
	}
}
