package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/schema"
)

func TestPipelineWithoutHTTPTriggerMethod(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := parse.LoadPipelines(ctx, "./pipelines/http_trigger_method.fp")
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

	httpTriggerWithoutMethod := triggers["local.trigger.http.trigger_without_method_block"]
	if httpTriggerWithoutMethod == nil {
		assert.Fail("trigger_without_method_block trigger not found")
		return
	}
	assert.Equal(true, *httpTriggerWithoutMethod.Enabled)

	httpTrigConfig, ok := httpTriggerWithoutMethod.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("trigger_without_method_block trigger is not a HTTP trigger")
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
	assert.Equal("synchronous", methodInfo.ExecutionMode)
}

func TestPipelineWithHTTPGetMethod(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := parse.LoadPipelines(ctx, "./pipelines/http_trigger_method.fp")
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

	httpTriggerWithoutMethod := triggers["local.trigger.http.trigger_with_get_method"]
	if httpTriggerWithoutMethod == nil {
		assert.Fail("trigger_with_get_method trigger not found")
		return
	}
	assert.Equal(true, *httpTriggerWithoutMethod.Enabled)

	httpTrigConfig, ok := httpTriggerWithoutMethod.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("trigger_with_get_method trigger is not a HTTP trigger")
		return
	}

	triggerMethods := httpTrigConfig.Methods
	assert.Equal(1, len(triggerMethods))

	methodInfo := triggerMethods["get"]
	assert.NotNil(methodInfo, "method 'get' not found")

	pipelineInfo := methodInfo.Pipeline.AsValueMap()
	assert.Equal("local.pipeline.simple_with_trigger", pipelineInfo[schema.AttributeTypeName].AsString())

	argsInfo, err := methodInfo.GetArgs(nil)
	assert.Nil(err)
	assert.NotNil(argsInfo)
	assert.Equal("one", argsInfo["param_one"])
	assert.Equal(2, argsInfo["param_two_int"])
	assert.Equal("synchronous", methodInfo.ExecutionMode)
}

func TestPipelineWithHTTPTriggerMethodMultiple(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := parse.LoadPipelines(ctx, "./pipelines/http_trigger_method.fp")
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

	httpTriggerWithoutMethod := triggers["local.trigger.http.trigger_with_multiple_method"]
	if httpTriggerWithoutMethod == nil {
		assert.Fail("trigger_with_multiple_method trigger not found")
		return
	}
	assert.Equal(true, *httpTriggerWithoutMethod.Enabled)

	httpTrigConfig, ok := httpTriggerWithoutMethod.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("trigger_with_multiple_method trigger is not a HTTP trigger")
		return
	}

	triggerMethods := httpTrigConfig.Methods
	assert.Equal(2, len(triggerMethods))

	methodInfo := triggerMethods["get"]
	assert.NotNil(methodInfo, "method 'get' not found")

	pipelineInfo := methodInfo.Pipeline.AsValueMap()
	assert.Equal("local.pipeline.simple_with_trigger", pipelineInfo[schema.AttributeTypeName].AsString())

	argsInfo, err := methodInfo.GetArgs(nil)
	assert.Nil(err)
	assert.NotNil(argsInfo)
	assert.Equal("one", argsInfo["param_one"])
	assert.Equal(3, argsInfo["param_two_int"])
	assert.Equal("synchronous", methodInfo.ExecutionMode)

	methodInfo = triggerMethods["post"]
	assert.NotNil(methodInfo, "method 'post' not found")

	pipelineInfo = methodInfo.Pipeline.AsValueMap()
	assert.Equal("local.pipeline.simple_with_trigger", pipelineInfo[schema.AttributeTypeName].AsString())

	argsInfo, err = methodInfo.GetArgs(nil)
	assert.Nil(err)
	assert.NotNil(argsInfo)
	assert.Equal("one", argsInfo["param_one"])
	assert.Equal(2, argsInfo["param_two_int"])
	assert.Equal("synchronous", methodInfo.ExecutionMode)
}

func TestPipelineWithHTTPTriggerPrecedence(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := parse.LoadPipelines(ctx, "./pipelines/http_trigger_method.fp")
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

	httpTriggerWithoutMethod := triggers["local.trigger.http.test_method_precedence"]
	if httpTriggerWithoutMethod == nil {
		assert.Fail("test_method_precedence trigger not found")
		return
	}
	assert.Equal(true, *httpTriggerWithoutMethod.Enabled)

	httpTrigConfig, ok := httpTriggerWithoutMethod.Config.(*resources.TriggerHttp)
	if !ok {
		assert.Fail("test_method_precedence trigger is not a HTTP trigger")
		return
	}

	triggerMethods := httpTrigConfig.Methods
	assert.Equal(1, len(triggerMethods))

	methodInfo := triggerMethods["post"]
	assert.Nil(methodInfo)

	methodInfo = triggerMethods["get"]
	assert.NotNil(methodInfo, "method 'get' not found")

	pipelineInfo := methodInfo.Pipeline.AsValueMap()
	assert.Equal("local.pipeline.simple_with_trigger", pipelineInfo[schema.AttributeTypeName].AsString())

	argsInfo, err := methodInfo.GetArgs(nil)
	assert.Nil(err)
	assert.NotNil(argsInfo)
	assert.Equal("one", argsInfo["param_one"])
	assert.Equal(3, argsInfo["param_two_int"])
	assert.Equal("synchronous", methodInfo.ExecutionMode)
}
