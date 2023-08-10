package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fpconfig"
	"github.com/turbot/flowpipe/internal/types"
)

func TestPipelineWithTrigger(t *testing.T) {
	assert := assert.New(t)

	fpParseContext, err := fpconfig.LoadFlowpipeConfig(context.TODO(), "./test_pipelines/with_trigger.fp")
	assert.Nil(err, "error found")

	pipelines := fpParseContext.PipelineHcls

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

	triggers := fpParseContext.TriggerHcls

	scheduleTrigger := triggers["my_hourly_trigger"]
	if scheduleTrigger == nil {
		assert.Fail("my_hourly_trigger trigger not found")
		return
	}

	st, ok := scheduleTrigger.(*types.TriggerSchedule)
	if !ok {
		assert.Fail("my_hourly_trigger trigger is not a schedule trigger")
		return
	}

	assert.Equal("5 * * * * *", st.Schedule)
}
