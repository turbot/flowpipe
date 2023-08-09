package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestPipelineWithTrigger(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/with_trigger.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["simple_with_trigger"] == nil {
		assert.Fail("simple_with_trigger pipeline not found")
		return
	}

	echoStep := pipelines["simple_with_trigger"].GetStep("echo.simple_echo")
	if echoStep == nil {
		assert.Fail("echo.simple_echo step not found")
		return
	}

	dependsOn := echoStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)
}
