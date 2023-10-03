package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestInvalidHttpTrigger(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_http_trigger.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `Unsupported argument: An argument named "if" is not expected here.`)
}

func TestInvalidStepAttribute(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_step_attribute.fp")
	assert.NotNil(err)
	assert.Contains(err.Error(), `Unsupported argument: An argument named "abc" is not expected here.`)
}
