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

func TestHttpStepLoad(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/http_step.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 4, "wrong number of pipelines")

	if pipelines["local.pipeline.http_step"] == nil {
		assert.Fail("http_step pipeline not found")
		return
	}

	pipelineHcl := pipelines["local.pipeline.http_step"]
	step := pipelineHcl.GetStep("http.send_to_slack")
	if step == nil {
		assert.Fail("http.send_to_slack step not found")
		return
	}

	stepInputs, err := step.GetInputs(nil)

	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")

	assert.Equal("https://myapi.com/vi/api/do-something", stepInputs[schema.AttributeTypeUrl], "wrong url")
	assert.Equal("post", stepInputs[schema.AttributeTypeMethod], "wrong method")
	assert.Equal("test", stepInputs[schema.AttributeTypeCaCertPem], "wrong cert")
	assert.Equal(true, stepInputs[schema.AttributeTypeInsecure], "wrong insecure")
	assert.Equal("{\"app\":\"flowpipe\",\"name\":\"turbie\"}", stepInputs[schema.AttributeTypeRequestBody], "wrong request_body")
	assert.Equal("flowpipe", stepInputs[schema.AttributeTypeRequestHeaders].(map[string]interface{})["User-Agent"], "wrong header")
}

func TestHttpStepLoadTimeoutUnresolved(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/http_step.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 4, "wrong number of pipelines")

	if pipelines["local.pipeline.http_step_timeout_unresolved"] == nil {
		assert.Fail("http_step_timeout_unresolved pipeline not found")
		return
	}

	pipelineHcl := pipelines["local.pipeline.http_step_timeout_unresolved"]
	step := pipelineHcl.GetStep("http.send_to_slack")
	if step == nil {
		assert.Fail("http.send_to_slack step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"timeout": cty.NumberIntVal(2000),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	stepInputs, err := step.GetInputs(evalContext)

	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")

	assert.Equal("https://myapi.com/vi/api/do-something", stepInputs[schema.AttributeTypeUrl], "wrong url")
	assert.Equal("post", stepInputs[schema.AttributeTypeMethod], "wrong method")
	assert.Equal(2000, stepInputs[schema.AttributeTypeTimeout], "wrong cert")
}

func TestHttpStepLoadTimeoutString(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/http_step.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 4, "wrong number of pipelines")

	if pipelines["local.pipeline.http_step_timeout_string"] == nil {
		assert.Fail("http_step_timeout_string pipeline not found")
		return
	}

	pipelineHcl := pipelines["local.pipeline.http_step_timeout_string"]
	step := pipelineHcl.GetStep("http.send_to_slack")
	if step == nil {
		assert.Fail("http.send_to_slack step not found")
		return
	}

	stepInputs, err := step.GetInputs(nil)

	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")

	assert.Equal("https://myapi.com/vi/api/do-something", stepInputs[schema.AttributeTypeUrl], "wrong url")
	assert.Equal("post", stepInputs[schema.AttributeTypeMethod], "wrong method")
	assert.Equal("2s", stepInputs[schema.AttributeTypeTimeout], "wrong cert")
}

func TestHttpStepLoadTimeoutStringUnresolved(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/http_step.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 4, "wrong number of pipelines")

	if pipelines["local.pipeline.http_step_timeout_string_unresolved"] == nil {
		assert.Fail("http_step_timeout_string_unresolved pipeline not found")
		return
	}

	pipelineHcl := pipelines["local.pipeline.http_step_timeout_string_unresolved"]
	step := pipelineHcl.GetStep("http.send_to_slack")
	if step == nil {
		assert.Fail("http.send_to_slack step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"timeout": cty.StringVal("2s"),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	stepInputs, err := step.GetInputs(evalContext)

	assert.Nil(err, "error found")
	assert.NotNil(stepInputs, "inputs not found")

	assert.Equal("https://myapi.com/vi/api/do-something", stepInputs[schema.AttributeTypeUrl], "wrong url")
	assert.Equal("post", stepInputs[schema.AttributeTypeMethod], "wrong method")
	assert.Equal("2s", stepInputs[schema.AttributeTypeTimeout], "wrong cert")
}
