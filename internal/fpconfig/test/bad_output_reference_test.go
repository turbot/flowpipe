package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestBadOutputReference(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/bad_output_reference.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `invalid depends_on 'echo.does_not_exist' - does not exist for pipeline local.pipeline`)
}
