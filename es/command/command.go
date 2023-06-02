package command

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type CommandHandler struct {
	// Command handlers can only send events, they are not even permitted access
	// to the CommandBus.
	EventBus *cqrs.EventBus
}
