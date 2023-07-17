package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/zclconf/go-cty/cty"
)

func TestSimpleForAndParam(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	pipelines, err := LoadPipelines(ctx, "./test_pipelines/for.fp")
	assert.Nil(err, "error found ")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["for_loop"] == nil {
		assert.Fail("for_loop pipeline not found")
		return
	}

	pipeline := pipelines["for_loop"]

	step := pipeline.GetStep("echo.no_for_each")
	if step == nil {
		assert.Fail("echo.no_for_each step not found")
		return
	}

	if step.GetForEach() != nil {
		assert.Fail("echo.no_for_each should not have a for_each")
		return
	}

	step = pipeline.GetStep("echo.text_1")

	if step == nil {
		assert.Fail("echo.text_1 step not found")
		return
	}

	objectVal := cty.ObjectVal(map[string]cty.Value{
		"users": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
		})})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = objectVal

	var output []string

	if step.GetForEach() == nil {
		assert.Fail("echo.text_1 should have a for_each")
		return
	}

	diag := gohcl.DecodeExpression(step.GetForEach(), evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}

	assert.Equal("foo bar", strings.Join(output, " "), "wrong output")

	textAttribute := step.GetUnresolvedAttributes()["text"]
	if textAttribute == nil {
		assert.Fail("text attribute not found")
	}

	eachVal := cty.ObjectVal(map[string]cty.Value{
		"value": cty.StringVal("foozball"),
	})

	var stringOutput string
	evalContext.Variables["each"] = eachVal

	diag = gohcl.DecodeExpression(textAttribute, evalContext, &stringOutput)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}

	assert.Equal("user if foozball", stringOutput, "wrong output")
}
