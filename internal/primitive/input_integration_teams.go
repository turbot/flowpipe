package primitive

import (
	"context"

	mst "github.com/atc0005/go-teams-notify/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

type InputIntegrationTeams struct {
	InputIntegrationBase
	WebhookUrl *string
}

func NewInputIntegrationTeams(base InputIntegrationBase) InputIntegrationTeams {
	return InputIntegrationTeams{InputIntegrationBase: base}
}

func (ip *InputIntegrationTeams) PostMessage(_ context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	output := modconfig.Output{}
	teams := mst.NewTeamsClient()
	err := teams.ValidateWebhook(*ip.WebhookUrl)
	if err != nil {
		return nil, err
	}

	msgCard, err := mc.TeamsMessage(ip, options)
	if err != nil {
		return nil, err
	}

	err = teams.Send(*ip.WebhookUrl, msgCard)
	return &output, err
}
