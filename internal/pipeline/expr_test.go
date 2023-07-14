package pipeline

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestExpression(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["text_expr"] == nil {
		assert.Fail("text_expr pipeline not found")
	}

	var output string
	expr := pipelines["text_expr"].Steps[1].GetUnresolvedAttributes()["text"]

	objectVal := cty.ObjectVal(map[string]cty.Value{
		"text": cty.ObjectVal(map[string]cty.Value{
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
	}
}

func TestExprFunc(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/expressions.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["expr_func"] == nil {
		assert.Fail("expr_func pipeline not found")
	}

	pipelineHcl := pipelines["expr_func"]
	step := pipelineHcl.GetStep("text.text_title")
	if step == nil {
		assert.Fail("text.text_title step not found")
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
