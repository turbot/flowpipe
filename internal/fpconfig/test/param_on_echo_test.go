package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestParamOnEcho(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/param_on_echo.fp")
	assert.Nil(err, "error found")

}
