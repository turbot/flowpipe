package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestStepOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/step_output.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.step_output"] == nil {
		assert.Fail("step_output pipeline not found")
		return
	}

	assert.Equal(2, len(pipelines["local.pipeline.step_output"].Steps), "wrong number of steps")

	startStep := pipelines["local.pipeline.step_output"].GetStep("echo.start_step")

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

	endStep := pipelines["local.pipeline.step_output"].GetStep("echo.end_step")
	if endStep == nil {
		assert.Fail("end_step not found")
		return
	}

	assert.Equal(1, len(endStep.GetDependsOn()))
	assert.Equal("echo.start_step", endStep.GetDependsOn()[0])
}
