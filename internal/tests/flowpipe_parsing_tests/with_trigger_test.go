package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/schema"
)

func TestPipelineWithTrigger(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := parse.LoadPipelines(ctx, "./pipelines/with_trigger.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.simple_with_trigger"] == nil {
		assert.Fail("simple_with_trigger pipeline not found")
		return
	}

	echoStep := pipelines["local.pipeline.simple_with_trigger"].GetStep("transform.simple_echo")
	if echoStep == nil {
		assert.Fail("transform.simple_echo step not found")
		return
	}

	dependsOn := echoStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)

	scheduleTrigger := triggers["local.trigger.schedule.my_hourly_trigger"]
	if scheduleTrigger == nil {
		assert.Fail("my_hourly_trigger trigger not found")
		return
	}
	assert.Equal(false, *scheduleTrigger.Enabled)

	st, ok := scheduleTrigger.Config.(*resources.TriggerSchedule)
	if !ok {
		assert.Fail("my_hourly_trigger trigger is not a schedule trigger")
		return
	}

	assert.Equal("5 * * * *", st.Schedule)

	scheduleTrigger = triggers["local.trigger.schedule.my_hourly_trigger_interval"]
	if scheduleTrigger == nil {
		assert.Fail("my_hourly_trigger_interval trigger not found")
		return
	}
	assert.Equal(true, *scheduleTrigger.Enabled)

	st, ok = scheduleTrigger.Config.(*resources.TriggerSchedule)
	if !ok {
		assert.Fail("my_hourly_trigger trigger is not a schedule trigger")
		return
	}

	assert.Equal("daily", st.Schedule)

	triggerWithArgs := triggers["local.trigger.schedule.trigger_with_args"]
	if triggerWithArgs == nil {
		assert.Fail("trigger_with_args trigger not found")
		return
	}
	assert.Nil(triggerWithArgs.Enabled)

	twa, ok := triggerWithArgs.Config.(*resources.TriggerSchedule)
	if !ok {
		assert.Fail("trigger_with_args trigger is not a schedule trigger")
		return
	}

	assert.NotNil(twa, "trigger_with_args trigger is nil")

	queryTrigger := triggers["local.trigger.query.query_trigger"]
	if queryTrigger == nil {
		assert.Fail("query_trigger trigger not found")
		return
	}

	qt, ok := queryTrigger.Config.(*resources.TriggerQuery)
	if !ok {
		assert.Fail("query_trigger trigger is not a query trigger")
		return
	}

	assert.Equal("access_key_id", qt.PrimaryKey)
	assert.Contains(qt.Sql, "where create_date < now() - interval")

	httpTriggerWithArgs := triggers["local.trigger.http.trigger_with_args"]
	if httpTriggerWithArgs == nil {
		assert.Fail("trigger_with_args trigger not found")
		return
	}
	assert.Equal(true, *httpTriggerWithArgs.Enabled)

	httpTrigConfig, ok := httpTriggerWithArgs.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("trigger_with_args trigger is not a HTTP trigger")
		return
	}

	triggerMethods := httpTrigConfig.Methods
	assert.Equal(1, len(triggerMethods))

	methodInfo := triggerMethods["post"]
	assert.NotNil(methodInfo, "method 'post' not found")

	pipelineInfo := methodInfo.Pipeline.AsValueMap()
	assert.Equal("local.pipeline.simple_with_trigger", pipelineInfo[schema.AttributeTypeName].AsString())

	argsInfo, err := methodInfo.GetArgs(nil)
	assert.Nil(err)
	assert.NotNil(argsInfo)
	assert.Equal("one", argsInfo["param_one"])
	assert.Equal(2, argsInfo["param_two_int"])

	queryTrigger = triggers["local.trigger.query.query_trigger_interval"]
	if queryTrigger == nil {
		assert.Fail("query_trigger_interval trigger not found")
		return
	}
	assert.Equal(true, *queryTrigger.Enabled)

	qt, ok = queryTrigger.Config.(*resources.TriggerQuery)
	if !ok {
		assert.Fail("query_trigger trigger is not a query trigger")
		return
	}

	assert.Equal("access_key_id", qt.PrimaryKey)
	assert.Contains(qt.Sql, "where create_date < now() - interval")
	assert.Equal("daily", qt.Schedule)

	triggerWithExecutionMode := triggers["local.trigger.http.trigger_with_execution_mode"]
	if triggerWithExecutionMode == nil {
		assert.Fail("trigger_with_execution_mode trigger not found")
		return
	}

	trig, ok := triggerWithExecutionMode.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("trigger_with_execution_mode trigger is not a http trigger")
		return
	}

	triggerMethods = trig.Methods
	assert.Equal(1, len(triggerMethods))

	methodInfo = triggerMethods["post"]
	assert.NotNil(methodInfo, "method 'post' not found")
	assert.Equal("synchronous", methodInfo.ExecutionMode)
}

func TestPipelineWithTriggerSelf(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	_, _, err := parse.LoadPipelines(ctx, "./pipelines/with_trigger_self.fp")
	assert.Nil(err, "error found")
}
