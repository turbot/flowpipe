package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestIf(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/if.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["if"] == nil {
		assert.Fail("if pipeline not found")
		return
	}

	step := pipelines["if"].GetStep("echo.text_1")

	if step == nil {
		assert.Fail("echo.text_1 step not found")
		return
	}

	ifExpr := step.GetUnresolvedAttributes()["if"]
	if ifExpr == nil {
		assert.Fail("if expression not found")
		return
	}
}
