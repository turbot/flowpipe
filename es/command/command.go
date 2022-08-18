package command

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type CommandHandler struct {
	CommandBus *cqrs.CommandBus
	EventBus   *cqrs.EventBus
}
