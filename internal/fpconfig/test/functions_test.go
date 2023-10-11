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

func TestFunctions(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/functions.fp")
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
