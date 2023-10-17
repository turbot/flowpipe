package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestNestedThreeLevelJsonencode(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/nested_three_levels_jsonencode.fp")
	assert.Nil(err)
	assert.NotNil(pipelines)

	assert.Equal(3, len(pipelines))
	found := false
	for _, s := range pipelines["local.pipeline.middle"].Steps {
		if s.GetName() == "echo_two" && s.GetType() == "echo" {
			dependsOn := s.GetDependsOn()
			assert.Equal(1, len(dependsOn))
			assert.Equal("pipeline.call_bottom", dependsOn[0])
			found = true
		}
	}

	assert.True(found)

}
