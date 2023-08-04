package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/pipeline"
)

func TestEmailStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, err := pipeline.LoadPipelines(context.TODO(), "./test_pipelines/email.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["email"] == nil {
		assert.Fail("query pipeline not found")
		return
	}

	step := pipelines["email"].GetStep("email.test_email")
	if step == nil {
		assert.Fail("email step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}

	assert.Equal([]string{"recipient@example.com"}, inputs["to"])
	assert.Equal("sender@example.com", inputs["from"])
	assert.Equal("sendercredential", inputs["sender_credential"])
	assert.Equal("smtp.example.com", inputs["host"])
	assert.Equal("587", inputs["port"])
	assert.Equal("Test email", inputs["subject"])
	assert.Equal("This is a test email", inputs["body"])
	assert.Equal("Flowpipe", inputs["sender_name"])
}
