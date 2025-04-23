package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
)

func TestNestedThreeLevelJsonencode(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/nested_three_levels_jsonencode.fp")
	assert.Nil(err)
	assert.NotNil(pipelines)

	assert.Equal(3, len(pipelines))
	found := false
	for _, s := range pipelines["local.pipeline.middle"].Steps {
		if s.GetName() == "echo_two" && s.GetType() == "transform" {
			dependsOn := s.GetDependsOn()
			assert.Equal(1, len(dependsOn))
			assert.Equal("pipeline.call_bottom", dependsOn[0])
			found = true
		}
	}

	assert.True(found)

}
