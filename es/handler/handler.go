package handler

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus *cqrs.CommandBus
}
