package pipeline_hcl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser/configschema"
)

func TestLoadPipelineDir(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := LoadPipelines(context.TODO(), "./simple")
	assert.Nil(err, "error found")

	assert.Greater(len(pipelines), 1, "wrong number of pipelines")

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["simple_http"], "pipeline not found")
		assert.Equal("simple_http", pipelines["simple_http"].Name, "wrong pipeline name")

		for _, step := range pipelines["simple_http"].ISteps {
			if step.GetName() == "my_step_1" {
				assert.Equal(configschema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
				assert.Equal("http://localhost:8081", step.GetInput()["url"], "wrong step input")
			}
		}
	}
}
