package pipeline_hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPipelineDir(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("for_loop_using_http_request_body_json", "for_loop_using_http_request_body_json")
}
