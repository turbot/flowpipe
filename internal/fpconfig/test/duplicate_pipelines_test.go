package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestDuplicatePipelines(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/duplicate_pipelines.fp")

	if err == nil {
		assert.Fail("expected error not found")
		return
	}

	assert.Contains(err.Error(), "Mod defines more than one resource named 'local.pipeline.pipeline_007'")
}
