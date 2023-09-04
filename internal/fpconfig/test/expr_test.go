package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/zclconf/go-cty/cty"
)

func TestExpression(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.text_expr"] == nil {
		assert.Fail("text_expr pipeline not found")
		return
	}

	var output string
	expr := pipelines["local.pipeline.text_expr"].Steps[1].GetUnresolvedAttributes()["text"]

	objectVal := cty.ObjectVal(map[string]cty.Value{
		"echo": cty.ObjectVal(map[string]cty.Value{
			"text_1": cty.ObjectVal(map[string]cty.Value{
				"text": cty.StringVal("hello"),
			}),
		}),
	})
	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["step"] = objectVal

	diag := gohcl.DecodeExpression(expr, evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}
	assert.Equal("bar hello baz", output, "wrong output")
}

func TestExprFunc(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.expr_func"] == nil {
		assert.Fail("expr_func pipeline not found")
		return
	}

	pipelineHcl := pipelines["local.pipeline.expr_func"]
	step := pipelineHcl.GetStep("echo.text_title")
	if step == nil {
		assert.Fail("echo.text_title step not found")
		return
	}

	stepInputs, err := step.GetInputs(nil)
	assert.Nil(err, "error found")
	assert.GreaterOrEqual(len(stepInputs), 1, "wrong number of inputs")

	textInput := stepInputs["text"]
	assert.NotNil(textInput, "text input not found")

	// test the title function is working as expected
	assert.Equal("Hello World", textInput, "wrong input format")
	assert.NotEqual("hello world", textInput, "wrong input format")
}

func TestExprWithinVariable(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.expr_within_text"] == nil {
		assert.Fail("expr_func pipeline not found")
	}

	pipelineHcl := pipelines["local.pipeline.expr_within_text"]
	step := pipelineHcl.GetStep("echo.text_title")
	if step == nil {
		assert.Fail("echo.text_title step not found")
	}

	// There's no unresolved variable, the function is just ${title("world")}
	assert.True(step.IsResolved(), "step should be resolved")

	stepInputs, err := step.GetInputs(nil)
	assert.Nil(err, "error found")
	assert.GreaterOrEqual(len(stepInputs), 1, "wrong number of inputs")

	textInput := stepInputs["text"]
	assert.NotNil(textInput, "text input not found")

	// test the title function is working as expected
	assert.Equal("Hello World", textInput, "wrong input format")
	assert.NotEqual("hello world", textInput, "wrong input format")
}

func TestExprDependAndFunction(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.expr_depend_and_function"] == nil {
		assert.Fail("expr_depend_and_function pipeline not found")
	}

	pipelineHcl := pipelines["local.pipeline.expr_depend_and_function"]
	stepOne := pipelineHcl.GetStep("echo.text_1")
	if stepOne == nil {
		assert.Fail("echo.text_1 step not found")
		return
	}

	assert.True(stepOne.IsResolved(), "step should be resolved")

	stepOneInput, err := stepOne.GetInputs(nil)
	assert.Nil(err)
	assert.Equal("foo", stepOneInput["text"])

	stepOneA := pipelineHcl.GetStep("echo.text_1_a")
	if stepOneA == nil {
		assert.Fail("echo.text_1_a step not found")
		return
	}

	assert.True(stepOneA.IsResolved(), "step should be resolved")

	stepOneAInput, err := stepOneA.GetInputs(nil)
	assert.Nil(err)

	// step_1_a has a title function on its text
	assert.Equal("Foo", stepOneAInput["text"])

	stepTwo := pipelineHcl.GetStep("echo.text_2")
	if stepTwo == nil {
		assert.Fail("echo.text_1 step not found")
		return
	}

	assert.False(stepTwo.IsResolved(), "step 2 should NOT be resolved")

	stepThree := pipelineHcl.GetStep("echo.text_3")
	if stepThree == nil {
		assert.Fail("text.text_3 step not found")
		return
	}

	assert.False(stepThree.IsResolved(), "step 3 should NOT be resolved")
}
