package command

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type CommandHandler struct {
	// Command handlers can only send events, they are not even permitted access
	// to the CommandBus.
	EventBus *FpEventBus
}

type FpEventBus struct {
	Eb *cqrs.EventBus
}

// Publish sends event to the event bus.
func (c FpEventBus) Publish(ctx context.Context, event interface{}) error {
	return c.Eb.Publish(ctx, event)
}
