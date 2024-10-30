package primitive

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/helpers"
)

type InputIntegrationSlack struct {
	InputIntegrationBase
	Token         *string
	SigningSecret *string
	WebhookUrl    *string
	Channel       *string
}

func NewInputIntegrationSlack(base InputIntegrationBase) InputIntegrationSlack {
	return InputIntegrationSlack{
		InputIntegrationBase: base,
	}
}

func (ip *InputIntegrationSlack) PostMessage(ctx context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*resources.Output, error) {
	var blocks slack.Blocks

	blocks, err := mc.SlackMessage(ip, options)
	if err != nil {
		return nil, err
	}

	output := resources.Output{}
	if !helpers.IsNil(ip.Token) && !helpers.IsNil(ip.Channel) {
		msgOption := slack.MsgOptionBlocks(blocks.BlockSet...)
		api := slack.New(*ip.Token)
		_, _, err = api.PostMessage(*ip.Channel, msgOption, slack.MsgOptionAsUser(false))
		return &output, err
	} else {
		wMsg := slack.WebhookMessage{Blocks: &blocks}
		err = slack.PostWebhook(*ip.WebhookUrl, &wMsg)
		return &output, err
	}
}
