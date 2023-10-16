package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestChildPipeline(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/child_pipeline.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.parent"] == nil {
		assert.Fail("parent pipeline not found")
		return
	}

	childPipelineStep := pipelines["local.pipeline.parent"].GetStep("pipeline.child_pipeline")
	if childPipelineStep == nil {
		assert.Fail("pipeline.child_pipeline step not found")
		return
	}

	dependsOn := childPipelineStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)

	// Unresolved attributes should be null at this stage, we have fully parsed child_pipeline.fp
	unresolvedAttributes := childPipelineStep.GetUnresolvedAttributes()
	assert.Equal(0, len(unresolvedAttributes))
}

func TestChildPipelineWithArgs(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/child_pipeline.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.child_step_with_args"] == nil {
		assert.Fail("child_step_with_args pipeline not found")
		return
	}

	childPipelineStep := pipelines["local.pipeline.child_step_with_args"].GetStep("pipeline.child_pipeline")
	if childPipelineStep == nil {
		assert.Fail("pipeline.child_pipeline step not found")
		return
	}

	dependsOn := childPipelineStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)

	// We have fully parsed the file, we should not have unresolved attributes
	unresolvedAttributes := childPipelineStep.GetUnresolvedAttributes()
	assert.Equal(0, len(unresolvedAttributes))
}
