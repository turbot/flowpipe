package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
)

func TestInvalidHttpTrigger(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_http_trigger.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `Unsupported argument: An argument named "if" is not expected here.`)
}

func TestInvalidStepAttribute(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_step_attribute.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `Unsupported argument: An argument named "abc" is not expected here.`)
}

func TestInvalidParams(t *testing.T) {
	assert := assert.New(t)

	_, _, err := misc.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_params.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `invalid property path: params.message_retention_duration`)
}
