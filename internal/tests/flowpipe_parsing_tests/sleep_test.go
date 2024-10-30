package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestSleepStepLoad(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/sleep.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_sleep"] == nil {
		assert.Fail("pipeline_with_sleep pipeline not found")
		return
	}
	pipelineHcl := pipelines["local.pipeline.pipeline_with_sleep"]

	// Step #1
	step := pipelineHcl.GetStep("sleep.sleep_duration_string_input")
	if step == nil {
		assert.Fail("sleep.sleep_duration_string_input step not found")
		return
	}
	stepInputs, err := step.GetInputs(nil)
	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")
	assert.Equal("5s", stepInputs[schema.AttributeTypeDuration], "wrong url")

	// Step #2
	step = pipelineHcl.GetStep("sleep.sleep_duration_integer_input")
	if step == nil {
		assert.Fail("sleep.sleep_duration_integer_input step not found")
		return
	}
	stepInputs, err = step.GetInputs(nil)
	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")
	assert.Equal(int(2000), stepInputs[schema.AttributeTypeDuration], "wrong url")
}

func TestSleepStepLoadUnresolved(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/sleep.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 2, "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_sleep_unresolved"] == nil {
		assert.Fail("pipeline_with_sleep_unresolved pipeline not found")
		return
	}
	pipelineHcl := pipelines["local.pipeline.pipeline_with_sleep_unresolved"]

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"duration_string":  cty.StringVal("3s"),
		"duration_integer": cty.NumberIntVal(3000),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	// Step #1
	step := pipelineHcl.GetStep("sleep.sleep_duration_string_input")
	if step == nil {
		assert.Fail("sleep.sleep_duration_string_input step not found")
		return
	}
	stepInputs, err := step.GetInputs(evalContext)
	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")
	assert.Equal("3s", stepInputs[schema.AttributeTypeDuration], "wrong url")

	// Step #2
	step = pipelineHcl.GetStep("sleep.sleep_duration_integer_input")
	if step == nil {
		assert.Fail("sleep.sleep_duration_integer_input step not found")
		return
	}
	stepInputs, err = step.GetInputs(evalContext)
	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")
	assert.Equal(3000, stepInputs[schema.AttributeTypeDuration], "wrong url")
}
