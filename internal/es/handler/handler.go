package handler

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus *FpCommandBus
}

type FpCommandBus struct {
	Cb *cqrs.CommandBus
}

// Send sends command to the command bus.
func (c FpCommandBus) Send(ctx context.Context, cmd interface{}) error {

	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//
	err := LogEventMessage(ctx, cmd)
	if err != nil {
		return err
	}
	return c.Cb.Send(ctx, cmd)
}

func LogEventMessage(ctx context.Context, cmd interface{}) error {

	commandEvent, ok := cmd.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	// executionLogger writes the event to a file
	executionLogger := fplog.ExecutionLogger(ctx, commandEvent.GetEvent().ExecutionID)
	executionLogger.Sugar().Infow("es", "event_type", commandEvent.HandlerName(), "payload", cmd)

	err := executionLogger.Sync()
	if err != nil {
		// logger.Error("failed to sync execution logger", "error", err)
		return err
	}

	return nil
}
