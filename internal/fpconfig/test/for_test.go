package pipeline_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/misc"
	"github.com/zclconf/go-cty/cty"
)

func TestSimpleForAndParam(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	pipelines, _, err := misc.LoadPipelines(ctx, "./test_pipelines/for.fp")

	if err != nil {
		assert.Fail("error found", err)
		return
	}

	if len(pipelines) == 0 {
		assert.Fail("pipelines is nil")
		return
	}

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.for_loop"] == nil {
		assert.Fail("for_loop pipeline not found")
		return
	}

	pipeline := pipelines["local.pipeline.for_loop"]

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

func TestParamsProcessing(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	pipelines, _, err := misc.LoadPipelines(ctx, "./test_pipelines/for.fp")
	assert.Nil(err, "error found ")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.for_loop"] == nil {
		assert.Fail("for_loop pipeline not found")
		return
	}

	pipeline := pipelines["local.pipeline.for_loop"]

	step := pipeline.GetStep("echo.text_1")
	if step == nil {
		assert.Fail("echo.text_1 step not found")
		return
	}

	if step.GetForEach() == nil {
		assert.Fail("echo.text_1 should have a for_each")
		return
	}

	variable := pipeline.Params["users"]
	if variable == nil {
		assert.Fail("users variable not found")
		return
	}

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}

	params := map[string]cty.Value{}
	for k, v := range pipeline.Params {
		params[k] = v.Default
	}

	evalContext.Variables["param"] = cty.ObjectVal(params)

	if err != nil {
		assert.Fail("found error")
		return
	}

	var output []string

	diag := gohcl.DecodeExpression(step.GetForEach(), evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}

	assert.Equal("jerry Janis Jimi", strings.Join(output, " "), "wrong output")

	// use Value function
	ctyOutput, diags := step.GetForEach().Value(evalContext)
	if diags.HasErrors() {
		assert.Fail("error in getting step value")
		return
	}

	assert.NotNil(ctyOutput, "cty output not nil")

	forEachCtyVals := []map[string]cty.Value{}
	if ctyOutput.Type().IsTupleType() {
		listVal := ctyOutput.AsValueSlice()
		assert.Equal(3, len(listVal), "wrong number of values")

		for _, v := range listVal {
			forEachCtyVals = append(forEachCtyVals, map[string]cty.Value{
				"value": v,
			})
		}
	} else {
		assert.Fail("cty output is not a list type")
	}

	expected := []string{"jerry", "Janis", "Jimi"}

	for i, v := range forEachCtyVals {
		evalContext.Variables["each"] = cty.ObjectVal(v)
		stepInput, err := step.GetInputs(evalContext)
		assert.Nil(err, "error getting step inputs")
		assert.Equal(stepInput["text"], "user if "+expected[i], "wrong input")
	}
}
