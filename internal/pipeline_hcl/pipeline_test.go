package pipeline_hcl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/types"
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
				assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
			}
		}
	}
}

func TestMarshallUnmarshal(t *testing.T) {
	assert := assert.New(t)
	pipelines, err := LoadPipelines(context.TODO(), "./simple")
	assert.Nil(err, "error found")

	assert.Greater(len(pipelines), 1, "wrong number of pipelines")

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["simple_http"], "pipeline not found")
		assert.Equal("simple_http", pipelines["simple_http"].Name, "wrong pipeline name")

		data, err := json.Marshal(pipelines["simple_http"])
		assert.Nil(err, "error found, can't marshall")

		var p types.PipelineHcl
		err = json.Unmarshal(data, &p)
		assert.Nil(err, "error found, can't unmarshall")

		found := false
		for _, step := range pipelines["simple_http"].ISteps {
			if step.GetName() == "my_step_1" {
				found = true
				assert.Equal(configschema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
				assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
			}
		}
		assert.True(found, "step not found")
	}

}
