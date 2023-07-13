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

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/expressions")
	assert.Nil(err, "error found")

	assert.Equal(len(pipelines), 1, "wrong number of pipelines")

	var output string
	expr := pipelines["text_expr"].Steps[1].GetUnresolvedAttributes()["text"]

	// ctyVal := cty.StringVal("hello")
	// fooVal := cty.StringVal("foo")

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
	// evalContext.Variables["step"] = fooVal

	diag := gohcl.DecodeExpression(expr, evalContext, &output)
	if diag.HasErrors() {
		t.Fatal(diag)
	}

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["text_expr"], "pipeline not found")
	}
}
