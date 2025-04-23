package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/zclconf/go-cty/cty"
)

func TestTransformStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/transform.fp")
	assert.Nil(err, "error found")
	assert.Equal(7, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_transform_step"] == nil {
		assert.Fail("pipeline_with_transform_step pipeline not found")
		return
	}

	step := pipelines["local.pipeline.pipeline_with_transform_step"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("hello world", inputs["value"])

	// Pipeline 2

	if pipelines["local.pipeline.pipeline_with_transform_step_unresolved"] == nil {
		assert.Fail("pipeline_with_transform_step_unresolved pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_transform_step_unresolved"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"random_text": cty.StringVal("hello world"),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("hello world", inputs["value"])

	// Pipeline 3

	if pipelines["local.pipeline.pipeline_with_transform_step_number_test"] == nil {
		assert.Fail("pipeline_with_transform_step_number_test pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_transform_step_number_test"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	inputs, err = step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal(100, inputs["value"])

	// Pipeline 4

	if pipelines["local.pipeline.pipeline_with_transform_step_number_test_unresolved"] == nil {
		assert.Fail("pipeline_with_transform_step_number_test_unresolved pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_transform_step_number_test_unresolved"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"random": cty.NumberIntVal(1000),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal(1000, inputs["value"])

	// Pipeline 5

	if pipelines["local.pipeline.pipeline_with_transform_step_string_list"] == nil {
		assert.Fail("pipeline_with_transform_step_string_list pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_transform_step_string_list"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"users": cty.ListVal([]cty.Value{
			cty.StringVal("brian"),
			cty.StringVal("freddie"),
			cty.StringVal("john"),
			cty.StringVal("roger"),
		}),
	})
	eachVal := cty.ObjectVal(map[string]cty.Value{
		"value": cty.StringVal("freddie"),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal
	evalContext.Variables["each"] = eachVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("user if freddie", inputs["value"])

	// Pipeline 6

	if pipelines["local.pipeline.pipeline_with_transform_step_number_list"] == nil {
		assert.Fail("pipeline_with_transform_step_number_list pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_transform_step_number_list"].GetStep("transform.transform_test")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"users": cty.ListVal([]cty.Value{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.NumberIntVal(3),
		}),
	})
	eachVal = cty.ObjectVal(map[string]cty.Value{
		"value": cty.NumberIntVal(3),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal
	evalContext.Variables["each"] = eachVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("counter set to 3", inputs["value"])

	// Pipeline 7

	if pipelines["local.pipeline.transform_step_for_map"] == nil {
		assert.Fail("transform_step_for_map pipeline not found")
		return
	}

	step = pipelines["local.pipeline.transform_step_for_map"].GetStep("transform.text_1")
	if step == nil {
		assert.Fail("transform step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"legends": cty.ObjectVal(map[string]cty.Value{
			"janis": cty.ObjectVal(map[string]cty.Value{
				"age":       cty.NumberFloatVal(27),
				"last_name": cty.StringVal("joplin"),
			}),
			"jimi": cty.ObjectVal(map[string]cty.Value{
				"age":       cty.NumberFloatVal(27),
				"last_name": cty.StringVal("hendrix"),
			}),
			"jerry": cty.ObjectVal(map[string]cty.Value{
				"age":       cty.NumberFloatVal(53),
				"last_name": cty.StringVal("garcia"),
			}),
		}),
	})
	eachVal = cty.ObjectVal(map[string]cty.Value{
		"key": cty.StringVal("jimi"),
		"value": cty.ObjectVal(map[string]cty.Value{
			"age":       cty.NumberFloatVal(27),
			"last_name": cty.StringVal("hendrix"),
		}),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal
	evalContext.Variables["each"] = eachVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("jimi hendrix was 27", inputs["value"])
}
