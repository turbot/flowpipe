package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestFunctions(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/functions.fp")
	assert.Nil(err, "error found")
}
