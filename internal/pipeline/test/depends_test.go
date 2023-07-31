package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestImplicitDependsIndex(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/depends.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["depends_index"] == nil {
		assert.Fail("depends_index pipeline not found")
		return
	}

	step := pipelines["depends_index"].GetStep("echo.echo_1")
	if step == nil {
		assert.Fail("echo.echo_1 step not found")
		return
	}

	dependsOn := step.GetDependsOn()
	assert.Contains(dependsOn, "sleep.sleep_1")
}

func TestImplicitDepends(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/depends.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["implicit_depends"] == nil {
		assert.Fail("implicit_depends pipeline not found")
		return
	}

	step := pipelines["implicit_depends"].GetStep("sleep.sleep_2")
	if step == nil {
		assert.Fail("sleep.sleep_2 step not found")
		return
	}

	dependsOn := step.GetDependsOn()
	assert.Contains(dependsOn, "sleep.sleep_1")
}

func TestExplicitDependsOnIndex(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/depends.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["explicit_depends_index"] == nil {
		assert.Fail("explicit_depends_index pipeline not found")
		return
	}

	step := pipelines["explicit_depends_index"].GetStep("echo.echo_1")
	if step == nil {
		assert.Fail("echo.echo_1 step not found")
		return
	}

	dependsOn := step.GetDependsOn()
	assert.Contains(dependsOn, "sleep.sleep_1")
}

func TestImplicitQueryDepends(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/query_depends.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["query"] == nil {
		assert.Fail("query pipeline not found")
		return
	}

	step := pipelines["query"].GetStep("echo.result")
	if step == nil {
		assert.Fail("echo.result step not found")
		return
	}

	dependsOn := step.GetDependsOn()
	assert.Contains(dependsOn, "query.query_1")
}
