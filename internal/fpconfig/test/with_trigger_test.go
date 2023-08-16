package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

func TestPipelineWithTrigger(t *testing.T) {
	assert := assert.New(t)

	fpParseContext, err := pipeline.LoadFlowpipeConfig(context.TODO(), "./test_pipelines/with_trigger.fp")
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

	st, ok := scheduleTrigger.(*pipeline.TriggerSchedule)
	if !ok {
		assert.Fail("my_hourly_trigger trigger is not a schedule trigger")
		return
	}

	assert.Equal("5 * * * *", st.Schedule)

	triggerWithArgs := triggers["trigger_with_args"]
	if triggerWithArgs == nil {
		assert.Fail("trigger_with_args trigger not found")
		return
	}

	twa, ok := triggerWithArgs.(*pipeline.TriggerSchedule)
	if !ok {
		assert.Fail("trigger_with_args trigger is not a schedule trigger")
		return
	}

	assert.Equal("one", twa.Args["param_one"])
	assert.Equal(2, twa.Args["param_two_int"])

	queryTrigger := triggers["query_trigger"]
	if queryTrigger == nil {
		assert.Fail("query_trigger trigger not found")
		return
	}

	qt, ok := queryTrigger.(*pipeline.TriggerQuery)
	if !ok {
		assert.Fail("query_trigger trigger is not a query trigger")
		return
	}

	assert.Equal("access_key_id", qt.PrimaryKey)
	assert.Len(qt.Events, 1)
	assert.Equal("insert", qt.Events[0])
	assert.Equal("one", qt.Args["param_one"])
	assert.Equal(2, qt.Args["param_two_int"])
	assert.Contains(qt.Sql, "where create_date < now() - interval")
}

func TestBadTriggerConfig(t *testing.T) {
	assert := assert.New(t)

	fpParseContext, err := pipeline.LoadFlowpipeConfig(context.TODO(), "./test_pipelines/invalid_trigger.fp")
	assert.NotNil(err, "should have some errors")

	diags := fpParseContext.Diags

	assert.True(diags.HasErrors())

	assert.Contains(diags[0].Subject.Filename, "invalid_trigger.fp")
	assert.Contains(diags[0].Summary, "Unsupported attribute; This object does not have an attribute named \"bad_pipeline\".")

	assert.Contains(diags[1].Subject.Filename, "invalid_trigger.fp")
	assert.Contains(diags[1].Summary, "Missing required argument")

	assert.Contains(diags[1].Subject.Filename, "invalid_trigger.fp")
	assert.Contains("Invalid cron expression: bad cron format", diags[2].Summary)

	assert.Contains(diags[1].Subject.Filename, "invalid_trigger.fp")
	assert.Contains("Invalid interval", diags[3].Summary)

}
