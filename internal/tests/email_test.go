package pipeline_test

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/tests/test_init"
	"testing"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestEmailStep(t *testing.T) {
	test_init.SetAppSpecificConstants()
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./test_pipelines/email.fp")
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
	assert.Equal("sender@example.com", inputs["smtp_username"])
	assert.Equal("sendercredential", inputs["smtp_password"])
	assert.Equal("smtp.example.com", inputs["host"])
	assert.Equal(int64(587), inputs["port"])
	assert.Equal("Test email", inputs["subject"])
	assert.Equal("This is a test email", inputs["body"])
	assert.Equal("Flowpipe", inputs["sender_name"])
}

func TestEmailStepWithParam(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./test_pipelines/email.fp")
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
		"transform": cty.ObjectVal(map[string]cty.Value{
			"email_body": cty.ObjectVal(map[string]cty.Value{
				"value": cty.StringVal("This is an email body"),
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
	assert.Equal("sender@example.com", inputs["smtp_username"])
	assert.Equal("sendercredential", inputs["smtp_password"])
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
	assert.Contains(dependsOn, "transform.email_body")
}
