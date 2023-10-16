package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestAllParam(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/all_param.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.all_param"]
	assert.NotNil(pipeline)

	// all steps must have unresolved attributes
	for _, step := range pipeline.Steps {
		// except echo bazz
		if step.GetName() == "echo_baz" {
			assert.Nil(step.GetUnresolvedAttributes()["text"])
		} else {
			assert.NotNil(step.GetUnresolvedAttributes()["text"])
		}
	}
}
