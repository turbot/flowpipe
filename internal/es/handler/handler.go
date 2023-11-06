package handler

import (
	"context"
	"fmt"

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

var (
	pipelineCancelled    = PipelineCanceled{}
	pipelineFailed       = PipelineFailed{}
	pipelineFinished     = PipelineFinished{}
	pipelineLoaded       = PipelineLoaded{}
	pipelinePaused       = PipelinePaused{}
	pipelinePlanned      = PipelinePlanned{}
	pipelineQueued       = PipelineQueued{}
	pipelineResumed      = PipelineResumed{}
	pipelineStarted      = PipelineStarted{}
	pipelineStepFinished = PipelineStepFinished{}
	pipelineStepQueued   = PipelineStepQueued{}
	pipelineStepStarted  = PipelineStepStarted{}
)

// Send sends command to the command bus.
func (c FpCommandBus) Send(ctx context.Context, cmd interface{}) error {

	// Unfortunately we need to save the event log *before* we sernd this command to Watermill. This mean we have to figure out what the
	// event_type is manually. By the time it goes into the Watermill bus, it's too late.
	//

	var eventType string
	switch cmd.(type) {
	case event.PipelineCanceled:
		eventType = pipelineCancelled.HandlerName()
	case event.PipelineFailed:
		eventType = pipelineFailed.HandlerName()
	case event.PipelineFinished:
		eventType = pipelineFinished.HandlerName()
	case event.PipelineLoaded:
		eventType = pipelineLoaded.HandlerName()
	case event.PipelinePaused:
		eventType = pipelinePaused.HandlerName()
	case event.PipelinePlanned:
		eventType = pipelinePlanned.HandlerName()
	case event.PipelineQueued:
		eventType = pipelineQueued.HandlerName()
	case event.PipelineResumed:
		eventType = pipelineResumed.HandlerName()
	case event.PipelineStarted:
		eventType = pipelineStarted.HandlerName()
	case event.PipelineStepFinished:
		eventType = pipelineStepFinished.HandlerName()
	case event.PipelineStepQueued:
		eventType = pipelineStepQueued.HandlerName()
	case event.PipelineStepStarted:
	default:
		return perr.BadRequestWithMessage(fmt.Sprintf("invalid command type %T", cmd))
	}

	err := LogEventMessage(ctx, eventType, cmd)
	if err != nil {
		return err
	}
	return c.Cb.Send(ctx, cmd)
}

func LogEventMessage(ctx context.Context, eventType string, cmd interface{}) error {

	payload := cmd.(map[string]interface{})

	// executionLogger writes the event to a file
	executionLogger := fplog.ExecutionLogger(ctx, payload[""])
	executionLogger.Sugar().Infow("es", "event_type", eventType, "payload", cmd)

	err := executionLogger.Sync()
	if err != nil {
		// logger.Error("failed to sync execution logger", "error", err)
		return err
	}

	return nil
}
