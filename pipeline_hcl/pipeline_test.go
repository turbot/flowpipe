package pipeline_hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPipelineDir(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("for_loop_using_http_request_body_json", "for_loop_using_http_request_body_json")

	pipelines, err := LoadWorkspacePipelines("./simple")
	assert.Nil(err, "error found")

	assert.Equal(1, len(pipelines), "wrong number of pipelines")

	if len(pipelines) == 1 {
		assert.NotNil(pipelines["simple_http"], "pipeline not found")
		assert.Equal("simple_http", pipelines["simple_http"].Name, "wrong pipeline name")
	}
}
