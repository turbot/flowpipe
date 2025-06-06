package pipeline_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/schema"
)

func TestLoadPipelineDir(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/simple")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(4, len(pipelines), "pipelines are not loaded correctly")

	assert.NotNil(pipelines["local.pipeline.simple_http"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http", pipelines["local.pipeline.simple_http"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http"].Steps), 3, "steps are not loaded correctly")

	for _, step := range pipelines["local.pipeline.simple_http"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("5s", step.GetInputs()["duration"], "wrong step input")
		}
		if step.GetName() == "send_it" {
			assert.Equal(schema.BlockTypePipelineStepEmail, step.GetType(), "wrong step type")
			// assert.Equal("victor@turbot.com", step.GetInputs()["to"], "wrong step input")
		}
	}
}

func SkipTestLoadPipelineDirRecursive(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/**/*.fp")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(7, len(pipelines), "pipelines are not loaded correctly")

	assert.NotNil(pipelines["local.pipeline.simple_http"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http", pipelines["local.pipeline.simple_http"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http"].Steps), 3, "steps are not loaded correctly")

	for _, step := range pipelines["local.pipeline.simple_http"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("5s", step.GetInputs()["duration"], "wrong step input")
		}
		if step.GetName() == "send_it" {
			assert.Equal(schema.BlockTypePipelineStepEmail, step.GetType(), "wrong step type")
			// assert.Equal("victor@turbot.com", step.GetInputs()["to"], "wrong step input")
		}
	}
}

func SkipTestLoadPipelineFromFileMatchesGlob(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/simple/simple*.fp")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(len(pipelines), 4, "pipelines are not loaded correctly")

	// Validate individual pipelines defined in the file

	// Pipeline 1
	assert.NotNil(pipelines["local.pipeline.simple_http"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http", pipelines["local.pipeline.simple_http"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http"].Steps), 3, "steps are not loaded correctly")

	for _, step := range pipelines["local.pipeline.simple_http"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("5s", step.GetInputs()["duration"], "wrong step input")
		}
		if step.GetName() == "send_it" {
			assert.Equal(schema.BlockTypePipelineStepEmail, step.GetType(), "wrong step type")
			// assert.Equal("victor@turbot.com", step.GetInputs()["to"], "wrong step input")
		}
	}

	// Pipeline 2
	assert.NotNil(pipelines["local.pipeline.simple_http_2"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http_2", pipelines["local.pipeline.simple_http_2"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http_2"].Steps), 1, "steps are not loaded correctly")
	for _, step := range pipelines["local.pipeline.simple_http_2"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
	}

	// Pipeline 3
	assert.NotNil(pipelines["local.pipeline.sleep_with_output"], "pipeline not found")
	assert.Equal("local.pipeline.sleep_with_output", pipelines["local.pipeline.sleep_with_output"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.sleep_with_output"].Steps), 1, "steps are not loaded correctly")
	for _, step := range pipelines["local.pipeline.sleep_with_output"].Steps {
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("1s", step.GetInputs()["duration"], "wrong step input")
		}
	}

	// Pipeline 4
	assert.NotNil(pipelines["local.pipeline.simple_http_file_2"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http_file_2", pipelines["local.pipeline.simple_http_file_2"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http_file_2"].Steps), 1, "steps are not loaded correctly")
	for _, step := range pipelines["local.pipeline.simple_http_file_2"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
	}
}

func TestLoadPipelineSpecificFile(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/simple/simple.fp")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(3, len(pipelines), "pipelines are not loaded correctly")

	// Validate individual pipelines defined in the file

	// Pipeline 1
	assert.NotNil(pipelines["local.pipeline.simple_http"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http", pipelines["local.pipeline.simple_http"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http"].Steps), 3, "steps are not loaded correctly")

	for _, step := range pipelines["local.pipeline.simple_http"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("5s", step.GetInputs()["duration"], "wrong step input")
		}
		if step.GetName() == "send_it" {
			assert.Equal(schema.BlockTypePipelineStepEmail, step.GetType(), "wrong step type")
			// assert.Equal("victor@turbot.com", step.GetInputs()["to"], "wrong step input")
		}
	}

	// Pipeline 2
	assert.NotNil(pipelines["local.pipeline.simple_http_2"], "pipeline not found")
	assert.Equal("local.pipeline.simple_http_2", pipelines["local.pipeline.simple_http_2"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.simple_http_2"].Steps), 1, "steps are not loaded correctly")
	for _, step := range pipelines["local.pipeline.simple_http_2"].Steps {
		if step.GetName() == "my_step_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
		}
	}

	// Pipeline 3
	assert.NotNil(pipelines["local.pipeline.sleep_with_output"], "pipeline not found")
	assert.Equal("local.pipeline.sleep_with_output", pipelines["local.pipeline.sleep_with_output"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.sleep_with_output"].Steps), 1, "steps are not loaded correctly")
	for _, step := range pipelines["local.pipeline.sleep_with_output"].Steps {
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("1s", step.GetInputs()["duration"], "wrong step input")
		}
	}
}

func TestSleepWithOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/sleep_with_output/sleep_with_output.fp")
	assert.Nil(err, "error found")

	assert.Equal(1, len(pipelines), "wrong number of pipelines")
	assert.Equal(1, len(pipelines["local.pipeline.sleep_with_output"].Steps), "steps are not loaded correctly")

	assert.NotNil(pipelines["local.pipeline.sleep_with_output"], "pipeline not found")
	assert.Equal("local.pipeline.sleep_with_output", pipelines["local.pipeline.sleep_with_output"].Name(), "wrong pipeline name")

	for _, step := range pipelines["local.pipeline.sleep_with_output"].Steps {
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			// assert.Equal("1s", step.GetInputs()["duration"], "wrong step input")
		}
	}
}

func TestLoadPipelineDepends(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/depends/depends.fp")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(len(pipelines), 1, "pipelines are not loaded correctly")

	assert.NotNil(pipelines["local.pipeline.http_and_sleep_depends"], "pipeline not found")
	assert.Equal("local.pipeline.http_and_sleep_depends", pipelines["local.pipeline.http_and_sleep_depends"].Name(), "wrong pipeline name")
	assert.Equal(len(pipelines["local.pipeline.http_and_sleep_depends"].Steps), 2, "steps are not loaded correctly")

	for _, step := range pipelines["local.pipeline.http_and_sleep_depends"].Steps {
		if step.GetName() == "http_1" {
			assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
			// assert.Equal("http://api.open-notify.org/astros.json", step.GetInputs()["url"], "wrong step input")
		}
		if step.GetName() == "sleep_1" {
			assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
			assert.Equal("http.http_1", step.GetDependsOn()[0], "wrong step depends on")
		}
	}
}

func TestMarshallUnmarshal(t *testing.T) {
	assert := assert.New(t)
	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/pipeline_dir/simple/simple.fp")
	assert.Nil(err, "error found")

	// Check the number of pipelines loaded
	assert.Equal(len(pipelines), 3, "pipelines are not loaded correctly")

	for pp := range pipelines {
		assert.NotNil(pipelines[pp], "pipeline not found")

		data, err := json.Marshal(pipelines[pp])
		assert.Nil(err, "error found, can't marshall")

		var p resources.Pipeline
		err = json.Unmarshal(data, &p)
		assert.Nil(err, "error found, can't unmarshall")

		if pp == "simple_http" {
			assert.Equal("local.pipeline.simple_http", p.Name(), "wrong pipeline name")
			assert.Equal(3, len(p.Steps), "steps are not loaded correctly")

			for _, step := range p.Steps {
				if step.GetName() == "my_step_1" {
					assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
					// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
				}
				if step.GetName() == "sleep_1" {
					assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
					// assert.Equal("5s", step.GetInputs()["duration"], "wrong step input")
				}
				if step.GetName() == "send_it" {
					assert.Equal(schema.BlockTypePipelineStepEmail, step.GetType(), "wrong step type")
					// assert.Equal("victor@turbot.com", step.GetInputs()["to"], "wrong step input")
				}
			}
		}

		if pp == "simple_http_2" {
			assert.Equal("local.pipeline.simple_http_2", p.Name(), "wrong pipeline name")
			assert.Equal(1, len(p.Steps), "steps are not loaded correctly")

			for _, step := range p.Steps {
				if step.GetName() == "my_step_1" {
					assert.Equal(schema.BlockTypePipelineStepHttp, step.GetType(), "wrong step type")
					// assert.Equal("http://localhost:8081", step.GetInputs()["url"], "wrong step input")
				}
			}
		}

		if pp == "sleep_with_output" {
			assert.Equal("local.pipeline.sleep_with_output", p.Name(), "wrong pipeline name")
			assert.Equal(1, len(p.Steps), "steps are not loaded correctly")

			for _, step := range p.Steps {
				if step.GetName() == "sleep_1" {
					assert.Equal(schema.BlockTypePipelineStepSleep, step.GetType(), "wrong step type")
					// assert.Equal("1s", step.GetInputs()["duration"], "wrong step input")
				}
			}
		}
	}
}
