package middleware

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/internal/es/event"
)

// This middleware writes the command and event to the jsonl event log file
func EventMiddleware(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				slog.Error("invalid log payload", "error", err)
				return nil, err
			}

			executionID := pe.Event.ExecutionID
			eventType := message.HandlerNameFromCtx(msg.Context())

			var payload map[string]interface{}
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				slog.Error("invalid log payload", "error", err)
				return nil, err
			}

			// TODO: Add global hook here based on execution id and perhaps the event type
			// TODO: maybe allow for regex on the execution id?
			// TODO: watch for registration and deregistration of the global hooks

			slog.Info("event", "execution_id", executionID, "event_type", eventType, "payload", payload, "event", pe.Event)

			return h(msg)
		}
	}
}
