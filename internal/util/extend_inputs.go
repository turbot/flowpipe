package util

import (
	"encoding/json"
	flowpipe2 "github.com/turbot/flowpipe/internal/resources"
	"log/slog"
	"strings"

	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/pipe-fittings/schema"
)

// TODO: refactor/tidy
// extendInputs is only relevant for the "input step". This is the best location we can find right now.
//
// We have tried to encapsulate this in the:
// 1) Step Definition's getInput() in pipe-fittings, but it needs the salt and we believe that it's not appropriate to use the salt in pipe-fittings.
// 2) In the primitive itself, but it's too late. We need this information in the Event for the remote CLI use case.
func ExtendInputs(executionId, pipelineExecutionId, stepExecutionId, stepName string, input flowpipe2.Input) flowpipe2.Input {
	stepType := strings.Split(stepName, ".")[0]
	switch stepType {
	case "input":
		var notifyMap any
		if notifyImpl, ok := input[schema.AttributeTypeNotifier].(flowpipe2.NotifierImpl); ok {
			// serialise notifyCty to json
			jsonData, err := json.Marshal(notifyImpl)
			if err != nil {
				slog.Error("Failed to marshal cty value", "error", err)
				return input
			}

			err = json.Unmarshal(jsonData, &notifyMap)
			if err != nil {
				slog.Error("Failed to unmarshal json data", "error", err)
				return input
			}
		}

		if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok || notifyMap != nil {
			if notifyMap == nil {
				notifyMap = notifier
			}

			if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
				for _, n := range notifies {
					if notify, ok := n.(map[string]any); ok {
						integration := notify["integration"].(map[string]any)
						integrationType := integration["type"].(string)
						switch integrationType {
						case schema.IntegrationTypeEmail, schema.IntegrationTypeHttp:
							formUrl, err := GetHttpFormUrl(executionId, pipelineExecutionId, stepExecutionId)
							if err != nil {
								slog.Error("Failed to get http form URL", "error", err)
							} else {
								input[constants.FormUrl] = formUrl
							}
							return input
						default:
							// slack, msteams, etc - do nothing
						}
					}
				}
			} else {
				formUrl, err := GetHttpFormUrl(executionId, pipelineExecutionId, stepExecutionId)
				if err != nil {
					slog.Error("Failed to get http form URL", "error", err)
				} else {
					input[constants.FormUrl] = formUrl
				}
				return input
			}
		}
		return input
	default:
		return input
	}
}
