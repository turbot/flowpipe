package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

func TestHttpStepLoad(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/http.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["http_step"] == nil {
		assert.Fail("http_step pipeline not found")
		return
	}

	pipelineHcl := pipelines["http_step"]
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
	assert.Equal(int64(2000), stepInputs[schema.AttributeTypeRequestTimeoutMs], "wrong request_timeout_ms")
	assert.Equal(true, stepInputs[schema.AttributeTypeInsecure], "wrong insecure")
	assert.Equal("{\"app\":\"flowpipe\",\"name\":\"turbie\"}", stepInputs[schema.AttributeTypeRequestBody], "wrong request_body")
	assert.Equal("flowpipe", stepInputs[schema.AttributeTypeRequestHeaders].(map[string]interface{})["User-Agent"], "wrong header")

	// stepInputsList, ok := stepInputs["list_text"].([]string)
	// if !ok {
	// 	assert.Fail("list_text input not found")
	// }
	// assert.Equal(stepInputsList[0], "foo", "wrong input format")
	// assert.Equal(stepInputsList[1], "bar", "wrong input format")
	// assert.Equal(stepInputsList[2], "baz", "wrong input format")

}
