package primitive

import (
	"context"
	"encoding/json"
	mst "github.com/atc0005/go-teams-notify/v2"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type InputIntegrationMsTeams struct {
	InputIntegrationBase
	IntegrationName string
	WebhookUrl      *string
}

func NewInputIntegrationMsTeams(base InputIntegrationBase, name string) InputIntegrationMsTeams {
	return InputIntegrationMsTeams{InputIntegrationBase: base, IntegrationName: name}
}

func (ip *InputIntegrationMsTeams) PostMessage(_ context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*flowpipe.Output, error) {
	output := flowpipe.Output{}
	teams := mst.NewTeamsClient()
	err := teams.ValidateWebhook(*ip.WebhookUrl)
	if err != nil {
		return nil, err
	}

	msgCard, err := mc.MsTeamsMessage(ip, options)
	if err != nil {
		return nil, err
	}

	err = teams.Send(*ip.WebhookUrl, msgCard)
	return &output, err
}

func (ip *InputIntegrationMsTeams) buildReturnPayload(valueString string, prompt string) (string, error) {
	salt, err := util.GetGlobalSalt()
	if err != nil {
		return "", err
	}
	hash, err := util.CalculateHash(ip.StepExecutionID, salt)
	if err != nil {
		return "", err
	}

	response := map[string]any{
		"value":                 valueString,
		"execution_id":          ip.ExecutionID,
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
		"prompt":                prompt,
		"step_execution_token":  hash,
	}
	jsonData, _ := json.Marshal(response)
	return string(jsonData), nil
}
