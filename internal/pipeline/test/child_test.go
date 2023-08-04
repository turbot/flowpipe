package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestChildPipeline(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/child_pipeline.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["parent"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	childPipelineStep := pipelines["parent"].GetStep("pipeline.child_pipeline")
	if childPipelineStep == nil {
		assert.Fail("pipeline.child_pipeline step not found")
		return
	}

	dependsOn := childPipelineStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)

	// Check if the unresolved attributes are correct, it should contain a reference to pipeline
	unresolvedAttributes := childPipelineStep.GetUnresolvedAttributes()
	assert.NotNil(unresolvedAttributes["pipeline"])
}
