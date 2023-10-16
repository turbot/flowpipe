package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

// Middleware to delay the PipelineStepStart command execution (for backoff purpose)
//
// TODO: make it generic?
func PipelineStepStartCommandDelayMiddlewareWithContext(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			logger := fplog.Logger(ctx)

			handlerName := message.HandlerNameFromCtx(msg.Context())
			if handlerName != "command.pipeline_step_start" {
				return h(msg)
			}

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			executionID := pe.Event.ExecutionID
			if executionID == "" {
				return nil, perr.InternalWithMessage("no execution_id found in payload")
			}

			var payload event.PipelineStepStart
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				logger.Error("invalid log payload", "error", err)
				return nil, err
			}

			logger.Info("CommandDelayMiddlewareWithContext", "handlerName", handlerName, "payload", payload)

			if payload.DelayMs == 0 {
				return h(msg)
			}

			waitTime := time.Millisecond * time.Duration(payload.DelayMs)

			select {
			case <-ctx.Done():
				return h(msg)
			case <-time.After(waitTime):
				// go on
			}

			//time.Sleep(waitTime)

			return h(msg)
		}
	}
}
