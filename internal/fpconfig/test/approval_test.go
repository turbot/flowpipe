package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/misc"
	"github.com/turbot/pipe-fittings/modconfig"
)

func TestApproval(t *testing.T) {
	assert := assert.New(t)

	mod, err := misc.LoadPipelinesReturningItsMod(context.TODO(), "./test_pipelines/approval.fp")
	assert.Nil(err)
	assert.NotNil(mod)

	assert.Equal(3, len(mod.ResourceMaps.Integrations))

	integration := mod.ResourceMaps.Integrations["local.integration.slack.my_slack_app"]
	if integration == nil {
		assert.Fail("Integration not found")
		return
	}

	assert.Equal("local.integration.slack.my_slack_app", integration.Name())
	assert.Equal("slack", integration.(*modconfig.SlackIntegration).Type)
	assert.Equal("xoxp-111111", *integration.(*modconfig.SlackIntegration).Token)
	assert.Equal("Q#$$#@#$$#W", *integration.(*modconfig.SlackIntegration).SigningSecret)

	integration = mod.ResourceMaps.Integrations["local.integration.email.email_integration"]
	if integration == nil {
		assert.Fail("Integration not found")
		return
	}

	assert.Equal("local.integration.email.email_integration", integration.Name())
	assert.Equal("email", integration.(*modconfig.EmailIntegration).Type)
	assert.Equal("foo bar baz", *integration.(*modconfig.EmailIntegration).SmtpHost)
	assert.Equal("bar foo baz", *integration.(*modconfig.EmailIntegration).DefaultSubject)

	pipeline := mod.ResourceMaps.Pipelines["local.pipeline.approval"]
	if pipeline == nil {
		assert.Fail("Pipeline not found")
		return
	}

	inputStep, ok := pipeline.Steps[0].(*modconfig.PipelineStepInput)
	if !ok {
		assert.Fail("Pipeline step not found")
		return
	}

	assert.Equal("input", inputStep.Name)
	assert.NotNil(inputStep.Notify)
	assert.Equal("foo", *inputStep.Notify.Channel)

	integrationLink := inputStep.Notify.Integration
	assert.NotNil(integrationLink)
	integrationMap := integrationLink.AsValueMap()
	assert.NotNil(integrationMap)
	assert.Equal("xoxp-111111", integrationMap["token"].AsString())

	assert.Equal("remove this after integrated", *inputStep.Token)

	inputsAfterEval, err := inputStep.GetInputs(&hcl.EvalContext{})
	// the notify should override the inline definition (the inline definition should not be there after integrated 2023)
	assert.Nil(err)

	assert.Equal("xoxp-111111", inputsAfterEval["token"].(string))
}
