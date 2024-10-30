package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestEmailStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/email.fp")
	assert.Nil(err, "error found")
	assert.Equal(2, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_with_email_step"] == nil {
		assert.Fail("pipeline_with_email_step pipeline not found")
		return
	}

	step := pipelines["local.pipeline.pipeline_with_email_step"].GetStep("email.valid_email")
	if step == nil {
		assert.Fail("email step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}

	assert.Equal("admiring.dijkstra@example.com", inputs[schema.AttributeTypeSmtpUsername])
	assert.Equal("abcdefghijklmnop", inputs[schema.AttributeTypeSmtpPassword])
	assert.Equal(int64(587), inputs[schema.AttributeTypePort])
	assert.Equal("smtp.gmail.com", inputs[schema.AttributeTypeHost])

	assert.Equal("sleepy.feynman@example.com", inputs[schema.AttributeTypeFrom])
	assert.Equal("Feynman", inputs[schema.AttributeTypeSenderName])

	if _, ok := inputs[schema.AttributeTypeTo].([]string); !ok {
		assert.Fail("attribute 'to' must be a list of strings")
	}

	recipients := inputs[schema.AttributeTypeTo].([]string)
	assert.Equal(2, len(recipients))
	assert.Equal("friendly.curie@example.com", recipients[0])
	assert.Equal("angry.kepler@example.com", recipients[1])

	assert.Equal("Flowpipe Test", inputs[schema.AttributeTypeSubject])
	assert.Equal("This is a test plaintext email body to validate the email step functionality.", inputs[schema.AttributeTypeBody])
	assert.Equal("text/plain", inputs[schema.AttributeTypeContentType])

	if _, ok := inputs[schema.AttributeTypeCc].([]string); !ok {
		assert.Fail("attribute 'cc' must be a list of strings")
	}

	ccRecipients := inputs[schema.AttributeTypeCc].([]string)
	assert.Equal(1, len(ccRecipients))
	assert.Equal("serene.turing@example.com", ccRecipients[0])

	if _, ok := inputs[schema.AttributeTypeBcc].([]string); !ok {
		assert.Fail("attribute 'bcc' must be a list of strings")
	}

	bccRecipients := inputs[schema.AttributeTypeBcc].([]string)
	assert.Equal(1, len(bccRecipients))
	assert.Equal("elastic.bassi@example.com", bccRecipients[0])

	// Pipeline 2

	if pipelines["local.pipeline.pipeline_with_unresolved_email_step_attributes"] == nil {
		assert.Fail("pipeline_with_unresolved_email_step_attributes pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_with_unresolved_email_step_attributes"].GetStep("email.valid_email")
	if step == nil {
		assert.Fail("email step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"smtp_username": cty.StringVal("admiring.dijkstra@example.com"),
		"smtp_password": cty.StringVal("abcdefghijklmnop"),
		"host":          cty.StringVal("smtp.gmail.com"),
		"port":          cty.NumberIntVal(587),
		"from":          cty.StringVal("sleepy.feynman@example.com"),
		"sender_name":   cty.StringVal("Feynman"),
		"subject":       cty.StringVal("Flowpipe Test"),
		"content_type":  cty.StringVal("text/plain"),
		"body":          cty.StringVal("This is a test plaintext email body to validate the email step functionality."),
		"to": cty.ListVal([]cty.Value{
			cty.StringVal("friendly.curie@example.com"),
			cty.StringVal("angry.kepler@example.com"),
		}),
		schema.AttributeTypeCc: cty.ListVal([]cty.Value{
			cty.StringVal("serene.turing@example.com"),
		}),
		"bcc": cty.ListVal([]cty.Value{
			cty.StringVal("elastic.bassi@example.com"),
		}),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}

	assert.Equal("admiring.dijkstra@example.com", inputs[schema.AttributeTypeSmtpUsername])
	assert.Equal("abcdefghijklmnop", inputs[schema.AttributeTypeSmtpPassword])
	assert.Equal(int64(587), inputs[schema.AttributeTypePort])
	assert.Equal("smtp.gmail.com", inputs[schema.AttributeTypeHost])

	assert.Equal("sleepy.feynman@example.com", inputs[schema.AttributeTypeFrom])
	assert.Equal("Feynman", inputs[schema.AttributeTypeSenderName])

	if _, ok := inputs[schema.AttributeTypeTo].([]string); !ok {
		assert.Fail("attribute 'to' must be a list of strings")
	}

	recipients = inputs[schema.AttributeTypeTo].([]string)
	assert.Equal(2, len(recipients))
	assert.Equal("friendly.curie@example.com", recipients[0])
	assert.Equal("angry.kepler@example.com", recipients[1])

	assert.Equal("Flowpipe Test", inputs[schema.AttributeTypeSubject])
	assert.Equal("This is a test plaintext email body to validate the email step functionality.", inputs[schema.AttributeTypeBody])
	assert.Equal("text/plain", inputs[schema.AttributeTypeContentType])

	if _, ok := inputs[schema.AttributeTypeCc].([]string); !ok {
		assert.Fail("attribute 'cc' must be a list of strings")
	}

	ccRecipients = inputs[schema.AttributeTypeCc].([]string)
	assert.Equal(1, len(ccRecipients))
	assert.Equal("serene.turing@example.com", ccRecipients[0])

	if _, ok := inputs[schema.AttributeTypeBcc].([]string); !ok {
		assert.Fail("attribute 'bcc' must be a list of strings")
	}

	bccRecipients = inputs[schema.AttributeTypeBcc].([]string)
	assert.Equal(1, len(bccRecipients))
	assert.Equal("elastic.bassi@example.com", bccRecipients[0])
}
