package primitive

import (
	"context"
	"errors"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

func TestMessageWithSlackNotifierUsingTokenNoChannelSet(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:  schema.IntegrationTypeSlack,
						schema.AttributeTypeToken: "xoxb-f4k3-t0k3n",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.NotNil(err)
	var fpErr perr.ErrorModel
	errors.As(err, &fpErr)
	assert.Contains(fpErr.Detail, "slack notifications require a channel when using token auth, channel was not set")
}

func TestMessageWithSlackNotifierUsingTokenChannelSetOnStep(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt:  "Test Prompt",
		schema.AttributeTypeType:    constants.InputTypeButton,
		schema.AttributeTypeChannel: "#step",
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:  schema.IntegrationTypeSlack,
						schema.AttributeTypeToken: "xoxb-f4k3-t0k3n",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithSlackNotifierUsingTokenChannelSetOnNotifier(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeChannel: "#notify",
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:  schema.IntegrationTypeSlack,
						schema.AttributeTypeToken: "xoxb-f4k3-t0k3n",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithSlackNotifierUsingTokenChannelSetOnIntegration(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:    schema.IntegrationTypeSlack,
						schema.AttributeTypeToken:   "xoxb-f4k3-t0k3n",
						schema.AttributeTypeChannel: "#integration",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithSlackNotifierUsingWebHookChannelNotSet(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:       schema.IntegrationTypeSlack,
						schema.AttributeTypeWebhookUrl: "https://fake-website.com/slack/webhook/url",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithEmailNotifierNoRecipients(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:     schema.IntegrationTypeEmail,
						schema.AttributeTypeSmtpHost: "smtp.email.com",
						schema.AttributeTypeFrom:     "example@email.com",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.NotNil(err)
	var fpErr perr.ErrorModel
	errors.As(err, &fpErr)
	assert.Contains(fpErr.Detail, "email notifications require recipients; one of 'to', 'cc' or 'bcc' need to be set")
}

func TestMessageWithEmailNotifierRecipientsOnStep(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeTo: []any{"bob@example.com", "other@example.com"},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:     schema.IntegrationTypeEmail,
						schema.AttributeTypeSmtpHost: "smtp.email.com",
						schema.AttributeTypeFrom:     "example@email.com",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithEmailNotifierRecipientsOnNotifier(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeCc: []any{"bob@example.com", "other@example.com"},
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:     schema.IntegrationTypeEmail,
						schema.AttributeTypeSmtpHost: "smtp.email.com",
						schema.AttributeTypeFrom:     "example@email.com",
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestMessageWithEmailNotifierRecipientsOnIntegration(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	step := NewMessagePrimitive("exec_123test", "pexec_456test", "sexec_789test", "pipeline.test", "input.test")
	input := resources.Input(map[string]any{
		schema.AttributeTypePrompt: "Test Prompt",
		schema.AttributeTypeType:   constants.InputTypeButton,
		schema.AttributeTypeOptions: []any{
			map[string]any{
				schema.AttributeTypeValue: "a",
			},
			map[string]any{
				schema.AttributeTypeValue: "b",
			},
		},
		schema.AttributeTypeNotifier: map[string]any{
			schema.AttributeTypeNotifies: []any{
				map[string]any{
					schema.AttributeTypeIntegration: map[string]any{
						schema.AttributeTypeType:     schema.IntegrationTypeEmail,
						schema.AttributeTypeSmtpHost: "smtp.email.com",
						schema.AttributeTypeFrom:     "example@email.com",
						schema.AttributeTypeTo:       []any{"bob@example.com", "other@example.com"},
					},
				},
			},
		},
	})

	err := step.ValidateInput(ctx, input)
	assert.Nil(err)
}

// TestMessageStepSlackMessagePlainText tests that SlackMessage uses PlainTextType by default
func TestMessageStepSlackMessagePlainText(t *testing.T) {
	assert := assert.New(t)

	creator := &MessageStepMessageCreator{
		Text:   "*bold* and _italic_ text with <https://example.com|link>",
		Mrkdwn: false,
	}

	blocks, err := creator.SlackMessage(nil, nil)
	assert.Nil(err)
	assert.Equal(1, len(blocks.BlockSet))

	section, ok := blocks.BlockSet[0].(*slack.SectionBlock)
	assert.True(ok, "expected SectionBlock")
	assert.Equal(slack.PlainTextType, section.Text.Type)
	assert.Equal("*bold* and _italic_ text with <https://example.com|link>", section.Text.Text)
}

// TestMessageStepSlackMessageMrkdwn tests that SlackMessage uses MarkdownType when mrkdwn=true
func TestMessageStepSlackMessageMrkdwn(t *testing.T) {
	assert := assert.New(t)

	creator := &MessageStepMessageCreator{
		Text:   "*bold* and _italic_ text with <https://example.com|link>",
		Mrkdwn: true,
	}

	blocks, err := creator.SlackMessage(nil, nil)
	assert.Nil(err)
	assert.Equal(1, len(blocks.BlockSet))

	section, ok := blocks.BlockSet[0].(*slack.SectionBlock)
	assert.True(ok, "expected SectionBlock")
	assert.Equal(slack.MarkdownType, section.Text.Type)
	assert.Equal("*bold* and _italic_ text with <https://example.com|link>", section.Text.Text)
}

// TestMessageStepSlackMessageMrkdwnComparison demonstrates the difference between PlainTextType and MarkdownType
// This test serves as evidence for the PR showing what each type produces
func TestMessageStepSlackMessageMrkdwnComparison(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name         string
		text         string
		mrkdwn       bool
		expectedType string
		description  string
	}{
		{
			name:         "PlainText_Bold",
			text:         "*bold text*",
			mrkdwn:       false,
			expectedType: slack.PlainTextType,
			description:  "Without mrkdwn, asterisks are literal: *bold text*",
		},
		{
			name:         "Mrkdwn_Bold",
			text:         "*bold text*",
			mrkdwn:       true,
			expectedType: slack.MarkdownType,
			description:  "With mrkdwn, asterisks create bold: bold text",
		},
		{
			name:         "PlainText_Italic",
			text:         "_italic text_",
			mrkdwn:       false,
			expectedType: slack.PlainTextType,
			description:  "Without mrkdwn, underscores are literal: _italic text_",
		},
		{
			name:         "Mrkdwn_Italic",
			text:         "_italic text_",
			mrkdwn:       true,
			expectedType: slack.MarkdownType,
			description:  "With mrkdwn, underscores create italic: italic text",
		},
		{
			name:         "PlainText_Link",
			text:         "<https://flowpipe.io|Flowpipe>",
			mrkdwn:       false,
			expectedType: slack.PlainTextType,
			description:  "Without mrkdwn, link syntax is literal: <https://flowpipe.io|Flowpipe>",
		},
		{
			name:         "Mrkdwn_Link",
			text:         "<https://flowpipe.io|Flowpipe>",
			mrkdwn:       true,
			expectedType: slack.MarkdownType,
			description:  "With mrkdwn, link syntax creates clickable link: Flowpipe",
		},
		{
			name:         "PlainText_Code",
			text:         "`code block`",
			mrkdwn:       false,
			expectedType: slack.PlainTextType,
			description:  "Without mrkdwn, backticks are literal: `code block`",
		},
		{
			name:         "Mrkdwn_Code",
			text:         "`code block`",
			mrkdwn:       true,
			expectedType: slack.MarkdownType,
			description:  "With mrkdwn, backticks create inline code formatting",
		},
		{
			name:         "PlainText_ComplexMessage",
			text:         "🚨 *Alert*: Check <https://example.com|dashboard> for _details_",
			mrkdwn:       false,
			expectedType: slack.PlainTextType,
			description:  "Complex message without mrkdwn shows raw formatting characters",
		},
		{
			name:         "Mrkdwn_ComplexMessage",
			text:         "🚨 *Alert*: Check <https://example.com|dashboard> for _details_",
			mrkdwn:       true,
			expectedType: slack.MarkdownType,
			description:  "Complex message with mrkdwn renders bold, links, and italic properly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			creator := &MessageStepMessageCreator{
				Text:   tc.text,
				Mrkdwn: tc.mrkdwn,
			}

			blocks, err := creator.SlackMessage(nil, nil)
			assert.Nil(err)
			assert.Equal(1, len(blocks.BlockSet))

			section, ok := blocks.BlockSet[0].(*slack.SectionBlock)
			assert.True(ok, "expected SectionBlock")
			assert.Equal(tc.expectedType, section.Text.Type, tc.description)
			assert.Equal(tc.text, section.Text.Text)

			// Log the block JSON for PR evidence
			t.Logf("Test: %s", tc.name)
			t.Logf("  Text: %s", tc.text)
			t.Logf("  Mrkdwn: %v", tc.mrkdwn)
			t.Logf("  Block Type: %s", section.Text.Type)
			t.Logf("  Description: %s", tc.description)
		})
	}
}
