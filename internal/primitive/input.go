package primitive

import (
	"context"
	"fmt"

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
}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {

	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Input must define a type")
	}

	if _, ok := i[schema.AttributeTypeType].(string); !ok {
		return perr.BadRequestWithMessage("Input type must be a string")
	}

	// TODO: validate type is one of button, text, select, multiselect, combo, multicombo
	// TODO: other validations
	// inputType := i[schema.AttributeTypeType].(string)
	//
	// switch inputType {
	// case string(IntegrationTypeSlack):
	// 	// Validate token
	// 	if i[schema.AttributeTypeToken] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a token")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeToken].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input token must be a string")
	// 	}
	//
	// 	// Validate channel
	// 	if i[schema.AttributeTypeChannel] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a channel")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeChannel].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input channel must be a string")
	// 	}
	//
	// 	// Validate the prompt
	// 	if i[schema.AttributeTypePrompt] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a prompt")
	// 	}
	// 	if _, ok := i[schema.AttributeTypePrompt].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input prompt must be a string")
	// 	}
	//
	// 	// Validate the slack type
	// 	if i[schema.AttributeTypeSlackType] == nil {
	// 		return perr.BadRequestWithMessage("Slack input must define a slack type")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeSlackType].(string); !ok {
	// 		return perr.BadRequestWithMessage("Slack input slack type must be a string")
	// 	}
	//
	// 	// Validate the options
	// 	var options []string
	// 	if i[schema.AttributeTypeOptions] == nil {
	// 		return perr.BadRequestWithMessage("Slack input options must define options")
	// 	}
	// 	if _, ok := i[schema.AttributeTypeOptions].([]string); ok {
	// 		options = i[schema.AttributeTypeOptions].([]string)
	// 	}
	// 	if _, ok := i[schema.AttributeTypeOptions].([]interface{}); ok {
	// 		for _, v := range i[schema.AttributeTypeOptions].([]interface{}) {
	// 			options = append(options, v.(string))
	// 		}
	// 	}
	// 	if len(options) == 0 {
	// 		return perr.BadRequestWithMessage("Slack input options must have at least one option")
	// 	}
	// case string(IntegrationTypeEmail):
	// }

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

					if channel, ok := notify[schema.AttributeTypeChannel].(string); ok {
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

					// TODO: Validate, make it generic
					var inputType, prompt string
					if it, ok := input[schema.AttributeTypeType].(string); ok {
						inputType = it
					}

					if p, ok := input[schema.AttributeTypePrompt].(string); ok {
						prompt = p
					}

					_, err := s.PostMessage(ctx, inputType, prompt, resOptions)
					if err != nil {
						return nil, err
					}
				case schema.IntegrationTypeWebform:
					// TODO: implement output
				case schema.IntegrationTypeEmail:
					e := NewInputIntegrationEmail(base)

					if formUrl, ok := input["webform_url"].(string); ok {
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

					if to, ok := notify[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					} else if to, ok := integration[schema.AttributeTypeTo].([]any); ok {
						for _, t := range to {
							e.To = append(e.To, t.(string))
						}
					}

					if cc, ok := notify[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					} else if cc, ok := integration[schema.AttributeTypeCc].([]any); ok {
						for _, c := range cc {
							e.Cc = append(e.Cc, c.(string))
						}
					}

					if bcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
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

					// TODO: Validate?
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

	return ip.execute(ctx, input, &InputStepMessageCreator{
		Prompt:    prompt,
		InputType: inputType,
	})
}

type MessageCreator interface {
	EmailMessage(*InputIntegrationEmail) (string, error)
	SlackMessage() string
}
