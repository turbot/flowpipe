package handler

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type EventHandler struct {
	CommandBus *cqrs.CommandBus
	EventBus   *cqrs.EventBus
}
