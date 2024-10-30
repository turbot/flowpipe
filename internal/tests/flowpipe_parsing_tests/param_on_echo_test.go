package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamOnEcho(t *testing.T) {
	assert := assert.New(t)

	_, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/param_on_echo.fp")
	assert.Nil(err, "error found")

}
