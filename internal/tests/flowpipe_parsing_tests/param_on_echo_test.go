package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
)

func TestParamOnEcho(t *testing.T) {
	assert := assert.New(t)

	_, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/param_on_echo.fp")
	assert.Nil(err, "error found")

}
