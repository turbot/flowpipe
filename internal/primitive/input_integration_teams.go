package primitive

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
)

type InputIntegrationTeams struct {
	InputIntegrationBase
	WebhookUrl *string
}

func NewInputIntegrationTeams(base InputIntegrationBase) InputIntegrationTeams {
	return InputIntegrationTeams{InputIntegrationBase: base}
}

func (ip *InputIntegrationTeams) PostMessage(ctx context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	// TODO: #TeamsIntegrationImplementation
	return nil, fmt.Errorf("teams integration PostMessage - NOT IMPLEMENTED")
}
