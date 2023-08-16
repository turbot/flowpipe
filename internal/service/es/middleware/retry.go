package middleware

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Custom retry middleware. To be used with Go Channel pub/sub. We need to be able to
// ack the message after the retry > max retry. Otherwise the message will be re-delivered and
// and we end up in an infinite loop situation
//
// The majority of this code is based on the Watermill Retry middleware.
type Retry struct {
	Ctx context.Context

	// MaxRetries is maximum number of times a retry will be attempted.
	MaxRetries int

	// InitialInterval is the first interval between retries. Subsequent intervals will be scaled by Multiplier.
	InitialInterval time.Duration
	// MaxInterval sets the limit for the exponential backoff of retries. The interval will not be increased beyond MaxInterval.
	MaxInterval time.Duration
	// Multiplier is the factor by which the waiting interval will be multiplied between retries.
	Multiplier float64
	// MaxElapsedTime sets the time limit of how long retries will be attempted. Disabled if 0.
	MaxElapsedTime time.Duration
	// RandomizationFactor randomizes the spread of the backoff times within the interval of:
	// [currentInterval * (1 - randomization_factor), currentInterval * (1 + randomization_factor)].
	RandomizationFactor float64

	// OnRetryHook is an optional function that will be executed on each retry attempt.
	// The number of the current retry is passed as retryNum,
	OnRetryHook func(retryNum int, delay time.Duration)
}

// Middleware function returns the Retry middleware.
func (r Retry) Middleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		producedMessages, err := h(msg)
		if err == nil {
			return producedMessages, nil
		}

		// Short circuit for now
		msg.Ack()
		return nil, err

		// 	// Check if the error is a Flowpipe error instance. If it is, check if it's retryable.
		// 	if flowpipeError, ok := err.(pcerr.ErrorModel); ok {
		// 		if !flowpipeError.Retryable {
		// 			// IMPORTANT: must do this
		// 			msg.Ack()
		// 			return nil, err
		// 		}
		// 	} else {
		// 		// All other errors are NOT retryable
		// 		msg.Ack()
		// 		return nil, err
		// 	}

		// 	expBackoff := backoff.NewExponentialBackOff()
		// 	expBackoff.InitialInterval = r.InitialInterval
		// 	expBackoff.MaxInterval = r.MaxInterval
		// 	expBackoff.Multiplier = r.Multiplier
		// 	expBackoff.MaxElapsedTime = r.MaxElapsedTime
		// 	expBackoff.RandomizationFactor = r.RandomizationFactor

		// 	ctx := msg.Context()
		// 	if r.MaxElapsedTime > 0 {
		// 		var cancel func()
		// 		ctx, cancel = context.WithTimeout(ctx, r.MaxElapsedTime)
		// 		defer cancel()
		// 	}

		// 	retryNum := 1
		// 	expBackoff.Reset()
		// retryLoop:
		// 	for {
		// 		waitTime := expBackoff.NextBackOff()
		// 		select {
		// 		case <-ctx.Done():
		// 			return producedMessages, err
		// 		case <-time.After(waitTime):
		// 			// go on
		// 		}

		// 		producedMessages, err = h(msg)
		// 		if err == nil {
		// 			return producedMessages, nil
		// 		}

		// 		logger := fplog.Logger(r.Ctx)
		// 		logger.Error("Error occurred, retrying", "error", err,
		// 			"retryNum", retryNum,
		// 			"maxRetries", r.MaxRetries,
		// 			"waitTime", waitTime,
		// 			"elapsedTime", expBackoff.GetElapsedTime())

		// 		if r.OnRetryHook != nil {
		// 			r.OnRetryHook(retryNum, waitTime)
		// 		}

		// 		retryNum++
		// 		if retryNum > r.MaxRetries {
		// 			// IMPORTANT: must do this
		// 			msg.Ack()
		// 			break retryLoop
		// 		}
		// 	}

		// 	return nil, err
	}
}
