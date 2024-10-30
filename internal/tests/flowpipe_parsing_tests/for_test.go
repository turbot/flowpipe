package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestSimpleForAndParam(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	pipelines, _, err := parse.LoadPipelines(ctx, "./pipelines/for.fp")

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

	step := pipeline.GetStep("transform.no_for_each")
	if step == nil {
		assert.Fail("transform.no_for_each step not found")
		return
	}

	if step.GetForEach() != nil {
		assert.Fail("transform.no_for_each should not have a for_each")
		return
	}

	step = pipeline.GetStep("transform.text_1")

	if step == nil {
		assert.Fail("transform.text_1 step not found")
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
		assert.Fail("transform.text_1 should have a for_each")
		return
	}

	diag := gohcl.DecodeExpression(step.GetForEach(), evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}

	assert.Equal("foo bar", strings.Join(output, " "), "wrong output")

	textAttribute := step.GetUnresolvedAttributes()["value"]
	if textAttribute == nil {
		assert.Fail("value attribute not found")
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

	pipelines, _, err := parse.LoadPipelines(ctx, "./pipelines/for.fp")
	assert.Nil(err, "error found ")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.for_loop"] == nil {
		assert.Fail("for_loop pipeline not found")
		return
	}

	pipeline := pipelines["local.pipeline.for_loop"]

	step := pipeline.GetStep("transform.text_1")
	if step == nil {
		assert.Fail("transform.text_1 step not found")
		return
	}

	if step.GetForEach() == nil {
		assert.Fail("transform.text_1 should have a for_each")
		return
	}

	variable := pipeline.GetParam("users")
	if variable == nil {
		assert.Fail("users variable not found")
		return
	}

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}

	params := map[string]cty.Value{}
	for _, v := range pipeline.Params {
		params[v.Name] = v.Default
	}

	evalContext.Variables["param"] = cty.ObjectVal(params)

	if err != nil {
		assert.Fail("found error")
		return
	}

	var output []string

	diag := gohcl.DecodeExpression(step.GetForEach(), evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression " + diag.Error())
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
		assert.Equal(stepInput["value"], "user if "+expected[i], "wrong input")
	}
}
