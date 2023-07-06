package pipeline

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

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/simple")
	assert.Nil(err, "error found")

	assert.Greater(len(pipelines), 1, "wrong number of pipelines")

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["simple_http"], "pipeline not found")
		assert.Equal("simple_http", pipelines["simple_http"].Name, "wrong pipeline name")

		for _, step := range pipelines["simple_http"].Steps {
			if step.GetName() == "my_step_1" {
				assert.Equal(configschema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
				assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
			}
		}
	}
}

func TestSleepWithOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/sleep_with_output")
	assert.Nil(err, "error found")

	assert.Equal(len(pipelines), 1, "wrong number of pipelines")

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["sleep_with_output"], "pipeline not found")
		assert.Equal("sleep_with_output", pipelines["sleep_with_output"].Name, "wrong pipeline name")
	}
}

func TestLoadPipelineDepends(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/depends_on")
	assert.Nil(err, "error found")

	assert.Greater(len(pipelines), 0, "wrong number of pipelines")

	if len(pipelines) > 0 {
		assert.NotNil(pipelines["http_and_sleep_depends"], "pipeline not found")
		assert.Equal("http_and_sleep_depends", pipelines["http_and_sleep_depends"].Name, "wrong pipeline name")

		for _, step := range pipelines["http_and_sleep_depends"].Steps {
			if step.GetName() == "sleep_1" {
				assert.Equal(configschema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
				assert.Equal("http.http_1", step.GetDependsOn()[0], "wrong step depends on")
			}
		}
	}
}

func TestLoadPipelineInvalidDepends(t *testing.T) {
	assert := assert.New(t)

	_, err := LoadPipelines(context.TODO(), "./test_pipelines/invalid_depends_on")
	assert.NotNil(err, "error not found")

	// TODO: need to improve the error here, need more context? sub-code?
	assert.Contains(err.Error(), "invalid depends_on", "wrong error message")
}

func TestMarshallUnmarshal(t *testing.T) {
	assert := assert.New(t)
	pipelines, err := LoadPipelines(context.TODO(), "./test_pipelines/simple")
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
		for _, step := range pipelines["simple_http"].Steps {
			if step.GetName() == "my_step_1" {
				found = true
				assert.Equal(configschema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
				assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
			}
		}
		assert.True(found, "step not found")
	}

}
