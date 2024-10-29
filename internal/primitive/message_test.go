package primitive

import (
	"context"
	"errors"
	"github.com/turbot/flowpipe/internal/resources"
	"testing"

	"github.com/stretchr/testify/assert"
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
