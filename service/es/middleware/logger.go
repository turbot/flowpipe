package middleware

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

// This middleware writes the command and event to the jsonl event log file
func LogEventMiddlewareWithContext(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			logger := fplog.Logger(ctx)

			logger.Trace("LogEventMiddlewareWithContext", "msg", msg)

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			executionID := pe.Event.ExecutionID
			if executionID == "" {
				return nil, fperr.InternalWithMessage("no execution_id found in payload")
			}

			var payload map[string]interface{}
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			payloadWithoutEvent := make(map[string]interface{})
			for key, value := range payload {
				if key == "event" {
					continue
				}
				payloadWithoutEvent[key] = value
			}
			logger.Debug("Event log", "createdAt", pe.Event.CreatedAt.Format("15:04:05.000"), "handlerNameFromCtx", message.HandlerNameFromCtx(msg.Context()), "payload", payloadWithoutEvent)

			// executionLogger writes the event to a file
			executionLogger := fplog.ExecutionLogger(ctx, executionID)
			executionLogger.Sugar().Infow("es", "event_type", message.HandlerNameFromCtx(msg.Context()), "payload", payload)
			defer func() {
				err := executionLogger.Sync()
				if err != nil {
					logger.Error("failed to sync execution logger", "error", err)
				}
			}()

			return h(msg)
		}
	}
}
