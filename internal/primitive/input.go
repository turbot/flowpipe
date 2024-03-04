package primitive

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/constants"
	"strings"

	"github.com/slack-go/slack"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type IntegrationType string

type Input struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
	PipelineName        string
	StepName            string
}

func NewInputPrimitive(executionId, pipelineExecutionId, stepExecutionId, pipelineName, stepName string) *Input {
	return &Input{
		ExecutionID:         executionId,
		PipelineExecutionID: pipelineExecutionId,
		StepExecutionID:     stepExecutionId,
		PipelineName:        pipelineName,
		StepName:            stepName,
	}
}

type InputIntegration interface {
	PostMessage(ctx context.Context, inputType string, prompt string, options []InputIntegrationResponseOption) (*modconfig.Output, error)
}

type InputIntegrationBase struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

func NewInputIntegrationBase(input *Input) InputIntegrationBase {
	return InputIntegrationBase{
		ExecutionID:         input.ExecutionID,
		PipelineExecutionID: input.PipelineExecutionID,
		StepExecutionID:     input.StepExecutionID,
	}
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

type InputIntegrationResponseOption struct {
	Label    *string
	Value    *string
	Selected *bool
	Style    *string
}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {
	// validate type
	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Input must define a type")
	}
	inputType, inputTypeIsString := i[schema.AttributeTypeType].(string)
	if !inputTypeIsString {
		return perr.BadRequestWithMessage("Input type must be a string")
	}
	if !constants.IsValidInputType(inputType) {
		return perr.BadRequestWithMessage(fmt.Sprintf("Input type '%s' is not supported", inputType))
	}

	// validate options
	switch inputType {
	case constants.InputTypeText:
		// text type doesn't require options, but don't fail if we have them, just ignore
	default:
		// ensure has at least 1 option
		options, hasOpts := i[schema.AttributeTypeOptions].([]any)
		if !hasOpts || len(options) == 0 {
			return perr.BadRequestWithMessage(fmt.Sprintf("Input type '%s' requires options, no options were defined", inputType))
		}

		// ensure all options have a value
		for i, o := range options {
			option := o.(map[string]any)
			if helpers.IsNil(option[schema.AttributeTypeValue]) {
				return perr.BadRequestWithMessage(fmt.Sprintf("option %d has no value specified", i))
			}
		}
	}

	err := ip.validateInputNotifier(i)
	if err != nil {
		return err // will already be perr
	}

	return nil
}

func (ip *Input) validateInputNotifier(i modconfig.Input) error {
	notifier := i[schema.AttributeTypeNotifier].(map[string]any)
	notifies := notifier[schema.AttributeTypeNotifies].([]any)
	for _, n := range notifies {
		notify := n.(map[string]any)
		integration := notify["integration"].(map[string]any)
		integrationType := integration["type"].(string)

		switch integrationType {
		case schema.IntegrationTypeHttp:
			// no additional validations required
		case schema.IntegrationTypeSlack:
			// if using token, we need to specify channel, webhook_url approach has a bound channel
			if integration[schema.AttributeTypeToken] != nil {
				if _, stepChannel := i[schema.AttributeTypeChannel].(string); !stepChannel {
					if _, notifyChannel := notify[schema.AttributeTypeChannel].(string); !notifyChannel {
						if _, integrationChannel := integration[schema.AttributeTypeChannel].(string); !integrationChannel {
							return perr.BadRequestWithMessage("slack notifications require a channel when using token auth, channel was not set")
						}
					}
				}
			}
		case schema.IntegrationTypeEmail:
			// ensure we have recipients, these can be to, cc or bcc but as optional at each layer need to ensure we have a target
			var recipients []string

			if to, ok := i[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			} else if to, ok := notify[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			} else if to, ok := integration[schema.AttributeTypeTo].([]any); ok {
				for _, t := range to {
					recipients = append(recipients, t.(string))
				}
			}

			if cc, ok := i[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			} else if cc, ok := notify[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			} else if cc, ok := integration[schema.AttributeTypeCc].([]any); ok {
				for _, c := range cc {
					recipients = append(recipients, c.(string))
				}
			}

			if bcc, ok := i[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			} else if bcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			} else if bcc, ok := integration[schema.AttributeTypeBcc].([]any); ok {
				for _, b := range bcc {
					recipients = append(recipients, b.(string))
				}
			}

			if len(recipients) == 0 {
				return perr.BadRequestWithMessage("email notifications require recipients; one of 'to', 'cc' or 'bcc' need to be set")
			}
		}
	}

	return nil
}

func (ip *Input) execute(ctx context.Context, input modconfig.Input, mc MessageCreator) (*modconfig.Output, error) {
	output := &modconfig.Output{}

	base := NewInputIntegrationBase(ip)

	var resOptions []InputIntegrationResponseOption

	if options, ok := input[schema.AttributeTypeOptions].([]any); ok {
		for _, o := range options {
			opt := o.(map[string]any)
			option := InputIntegrationResponseOption{}
			if l, ok := opt[schema.AttributeTypeLabel].(string); ok {
				option.Label = &l
			}
			if v, ok := opt[schema.AttributeTypeValue].(string); ok {
				option.Value = &v
				if helpers.IsNil(option.Label) {
					option.Label = &v
				}
			}
			if s, ok := opt[schema.AttributeTypeSelected].(bool); ok {
				option.Selected = &s
			}
			if s, ok := opt[schema.AttributeTypeStyle].(string); ok {
				option.Style = &s
			}
			resOptions = append(resOptions, option)
		}
	}

	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			for _, n := range notifies {
				notify := n.(map[string]any)
				integration := notify["integration"].(map[string]any)
				integrationType := integration["type"].(string)

				switch integrationType {
				case schema.IntegrationTypeSlack:
					s := NewInputIntegrationSlack(base)

					// Three ways to set the channel, in order of precedence
					if channel, ok := input[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					} else if channel, ok := notify[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					} else if channel, ok := integration[schema.AttributeTypeChannel].(string); ok {
						s.Channel = &channel
					}

					if tkn, ok := integration[schema.AttributeTypeToken].(string); ok {
						s.Token = &tkn
					}
					if ss, ok := integration[schema.AttributeTypeSigningSecret].(string); ok {
						s.SigningSecret = &ss
					}
					if wu, ok := integration[schema.AttributeTypeWebhookUrl].(string); ok {
						s.WebhookUrl = &wu
					}

					_, err := s.PostMessage(ctx, mc, resOptions)
					if err != nil {
						return nil, err
					}
				case schema.IntegrationTypeHttp:
					// No output needs to be rendered here for HTTP step. The console output is rendered by the Event printer, it does the right thing there too.
				case schema.IntegrationTypeEmail:
					e := NewInputIntegrationEmail(base)

					if formUrl, ok := input["form_url"].(string); ok {
						e.FormUrl = formUrl
					}

					if host, ok := integration[schema.AttributeTypeSmtpHost].(string); ok {
						e.Host = &host
					}
					if port, ok := integration[schema.AttributeTypeSmtpPort].(int64); ok {
						e.Port = &port
					} else if port, ok := integration[schema.AttributeTypeSmtpPort].(float64); ok {
						intPort := int64(port)
						e.Port = &intPort
					}
					if sPort, ok := integration[schema.AttributeTypeSmtpsPort].(int64); ok {
						e.SecurePort = &sPort
					} else if sPort, ok := integration[schema.AttributeTypeSmtpsPort].(float64); ok {
						intPort := int64(sPort)
						e.SecurePort = &intPort
					}
					if tls, ok := integration[schema.AttributeTypeSmtpTls].(string); ok {
						e.Tls = &tls
					}
					if from, ok := integration[schema.AttributeTypeFrom].(string); ok {
						e.From = from
					}

					if to, ok := input[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					} else if to, ok := notify[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					} else if to, ok := integration[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					}

					if cc, ok := input[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					} else if cc, ok := notify[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					} else if cc, ok := integration[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					}

					if bcc, ok := input[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					} else if bcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					} else if bcc, ok := integration[schema.AttributeTypeBcc].([]any); ok {
						for _, b := range bcc {
							e.Bcc = append(e.Bcc, b.(string))
						}
					}

					if sub, ok := input[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					} else if sub, ok := notify[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					} else if sub, ok := integration[schema.AttributeTypeSubject].(string); ok {
						e.Subject = sub
					}

					if u, ok := integration[schema.AttributeTypeSmtpUsername].(string); ok {
						e.User = &u
					}
					if p, ok := integration[schema.AttributeTypeSmtpPassword].(string); ok {
						e.Pass = &p
					}

					out, err := e.PostMessage(ctx, mc, resOptions)
					if err != nil {
						return nil, err
					}
					if out != nil {
						output = out
					}

				default:
					return nil, perr.InternalWithMessage(fmt.Sprintf("integration type %s not yet implemented", integrationType))
				}
			}
		}
	}

	return output, nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	var inputType, prompt string
	if it, ok := input[schema.AttributeTypeType].(string); ok {
		inputType = it
	}

	if p, ok := input[schema.AttributeTypePrompt].(string); ok {
		prompt = p
	}

	stepName := strings.Split(ip.StepName, ".")[len(strings.Split(ip.StepName, "."))-1]

	return ip.execute(ctx, input, &InputStepMessageCreator{
		Prompt:    prompt,
		InputType: inputType,
		StepName:  stepName,
	})
}

type MessageCreator interface {
	EmailMessage(*InputIntegrationEmail, []InputIntegrationResponseOption) (string, error)
	SlackMessage(*InputIntegrationSlack, []InputIntegrationResponseOption) (slack.Blocks, error)
}
