package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/zclconf/go-cty/cty"
)

func TestEmailStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/email.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.email"] == nil {
		assert.Fail("email pipeline not found")
		return
	}

	step := pipelines["local.pipeline.email"].GetStep("email.test_email")
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
	assert.Equal(int64(587), inputs["port"])
	assert.Equal("Test email", inputs["subject"])
	assert.Equal("This is a test email", inputs["body"])
	assert.Equal("Flowpipe", inputs["sender_name"])
}

func TestEmailStepWithParam(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/email.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.subscribe"] == nil {
		assert.Fail("pipeline not found")
		return
	}

	step := pipelines["local.pipeline.subscribe"].GetStep("email.send_it")
	if step == nil {
		assert.Fail("email step not found")
		return
	}

	var output string
	expr := pipelines["local.pipeline.subscribe"].Steps[1].GetUnresolvedAttributes()["body"]

	objectVal := cty.ObjectVal(map[string]cty.Value{
		"echo": cty.ObjectVal(map[string]cty.Value{
			"email_body": cty.ObjectVal(map[string]cty.Value{
				"text": cty.StringVal("This is an email body"),
			}),
		}),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["step"] = objectVal

	inputs, err := step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}

	assert.Contains(inputs["to"], "recipient@example.com")
	assert.Equal("sender@example.com", inputs["from"])
	assert.Equal("sendercredential", inputs["sender_credential"])
	assert.Equal("smtp.example.com", inputs["host"])
	assert.Equal(int64(587), inputs["port"])
	assert.Equal("You have been subscribed", inputs["subject"])

	diag := gohcl.DecodeExpression(expr, evalContext, &output)
	if diag.HasErrors() {
		assert.Fail("error decoding expression")
		return
	}
	assert.Equal("This is an email body", output, "wrong output")

	dependsOn := step.GetDependsOn()
	assert.Contains(dependsOn, "echo.email_body")
}

func TestEmailStepInvalidPortFormat(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_email_port.fp")
	assert.NotNil(err, "error found")

	assert.Contains(err.Error(), "Unable to convert port into integer")
}

func TestEmailStepInvalidRecipient(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/invalid_email_recipient.fp")
	assert.NotNil(err, "error found")
	assert.Contains(err.Error(), "Unable to parse to attribute to string slice")
}
