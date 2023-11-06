package command

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
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
	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, event)
	if err != nil {
		return err
	}

	return c.Eb.Publish(ctx, event)
}

func LogEventMessage(ctx context.Context, evt interface{}) error {

	commandEvent, ok := evt.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	// executionLogger writes the event to a file
	executionLogger := fplog.ExecutionLogger(ctx, commandEvent.GetEvent().ExecutionID)
	executionLogger.Sugar().Infow("es", "event_type", commandEvent.HandlerName(), "payload", evt)

	err := executionLogger.Sync()
	if err != nil {
		// logger.Error("failed to sync execution logger", "error", err)
		return err
	}

	return nil
}
