package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/step_output.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.step_output"] == nil {
		assert.Fail("step_output pipeline not found")
		return
	}

	assert.Equal(2, len(pipelines["local.pipeline.step_output"].Steps), "wrong number of steps")

	startStep := pipelines["local.pipeline.step_output"].GetStep("transform.start_step")

	startStepOutputConfig := startStep.GetOutputConfig()
	if startStepOutputConfig == nil {
		assert.Fail("output config not found")
	}

	startOutput := startStepOutputConfig["start_output"]
	if startOutput == nil {
		assert.Fail("start_output not found")
		return
	}

	assert.Equal("bar", startOutput.Value)

	endStep := pipelines["local.pipeline.step_output"].GetStep("transform.end_step")
	if endStep == nil {
		assert.Fail("end_step not found")
		return
	}

	assert.Equal(1, len(endStep.GetDependsOn()))
	assert.Equal("transform.start_step", endStep.GetDependsOn()[0])
}
