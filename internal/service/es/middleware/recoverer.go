package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

// Holds the recovered panic's error along with the stacktrace.
type RecoveredPanicError struct {
	V          interface{}
	Stacktrace string
}

func (p RecoveredPanicError) Error() string {
	return fmt.Sprintf("panic occurred: %#v, stacktrace: \n%s", p.V, p.Stacktrace)
}

// Recover from Go panic middleware. Based on Watermill Recoverer middleware.
//
// The panic will be wrapped in a Flowpipe Error and set as a fatal error (non-retryable).
func PanicRecovererMiddleware(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (messages []*message.Message, err error) {
			panicked := true

			defer func() {
				if r := recover(); r != nil || panicked {
					logger := fplog.Logger(ctx)
					logger.Error("Recovered from panic", "error", err)
					recoveredPanicErr := RecoveredPanicError{V: r, Stacktrace: string(debug.Stack())}

					// Flowpipe error by default is not retryable
					internalErr := perr.Internal(recoveredPanicErr)
					err = internalErr

					// Must ack here otherwise Watermill will go to an infinite loop

					// TODO: how do we retry? Should we do this in the router / Watermill? Or should we handle it in Flowpipe ES?
					msg.Ack()
				}
			}()

			messages, err = h(msg)
			panicked = false
			return messages, err
		}
	}
}
