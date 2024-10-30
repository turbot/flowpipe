package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestFunctionStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/function.fp")
	assert.Nil(err, "error found")
	assert.Equal(4, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.function_step_test"] == nil {
		assert.Fail("function_step_test pipeline not found")
		return
	}

	step := pipelines["local.pipeline.function_step_test"].GetStep("function.my_func")
	if step == nil {
		assert.Fail("function step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("./my-function", inputs[schema.AttributeTypeSource])
	assert.Equal("nodejs", inputs[schema.AttributeTypeRuntime])
	assert.Equal("my_file.my_handler", inputs[schema.AttributeTypeHandler])
	assert.Equal(10, inputs[schema.AttributeTypeTimeout])

	// Pipeline 2

	if pipelines["local.pipeline.function_step_test_with_param"] == nil {
		assert.Fail("function_step_test_with_param pipeline not found")
		return
	}

	step = pipelines["local.pipeline.function_step_test_with_param"].GetStep("function.my_func")
	if step == nil {
		assert.Fail("function step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"src":     cty.StringVal("./my-function"),
		"runtime": cty.StringVal("nodejs"),
		"timeout": cty.NumberIntVal(10),
		"handler": cty.StringVal("my_file.my_handler"),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("./my-function", inputs[schema.AttributeTypeSource])
	assert.Equal("nodejs", inputs[schema.AttributeTypeRuntime])
	assert.Equal("my_file.my_handler", inputs[schema.AttributeTypeHandler])
	assert.Equal(10, inputs[schema.AttributeTypeTimeout])

	// Pipeline 3

	if pipelines["local.pipeline.function_step_test_string_timeout"] == nil {
		assert.Fail("function_step_test_string_timeout pipeline not found")
		return
	}

	step = pipelines["local.pipeline.function_step_test_string_timeout"].GetStep("function.my_func")
	if step == nil {
		assert.Fail("function step not found")
		return
	}

	inputs, err = step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("./my-function", inputs[schema.AttributeTypeSource])
	assert.Equal("nodejs", inputs[schema.AttributeTypeRuntime])
	assert.Equal("my_file.my_handler", inputs[schema.AttributeTypeHandler])
	assert.Equal("10s", inputs[schema.AttributeTypeTimeout])

	// Pipeline 4

	if pipelines["local.pipeline.function_step_test_string_timeout_with_param"] == nil {
		assert.Fail("function_step_test_string_timeout_with_param pipeline not found")
		return
	}

	step = pipelines["local.pipeline.function_step_test_string_timeout_with_param"].GetStep("function.my_func")
	if step == nil {
		assert.Fail("function step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"src":     cty.StringVal("./my-function"),
		"runtime": cty.StringVal("nodejs"),
		"timeout": cty.StringVal("10s"),
		"handler": cty.StringVal("my_file.my_handler"),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("./my-function", inputs[schema.AttributeTypeSource])
	assert.Equal("nodejs", inputs[schema.AttributeTypeRuntime])
	assert.Equal("my_file.my_handler", inputs[schema.AttributeTypeHandler])
	assert.Equal("10s", inputs[schema.AttributeTypeTimeout])
}
