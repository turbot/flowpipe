package middleware

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PlannerControl struct {
	ctx context.Context
	mu  sync.Mutex
	// queuedCalls map[string]string

	// executionChannel map[string]chan interface{}
}

func NewPlannerControl(ctx context.Context) *PlannerControl {
	return &PlannerControl{
		ctx: ctx,
		// queuedCalls:      make(map[string]string),
		// executionChannel: map[string]chan interface{}{},
	}
}

func (p *PlannerControl) Middleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := fplog.Logger(p.ctx)

		eventName := msg.Metadata.Get("name")
		// // if eventName != "event.PipelinePlan" && eventName != "event.PipelinePlanned" && eventName != "event.PipelineFinished" && eventName != "event.PipelineCanceled" && eventName != "event.PipelineFailed" && eventName != "event.PipelineFinished" {
		// // 	return h(msg)
		// // }

		// logger.Info("Planner control middleware", "msg", msg)

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

		// if eventName == "event.PipelinePlan" {
		// 	if p.queuedCalls[executionID] != "" {
		// 		logger.Warn("XXXX execution ID is already queued", "execution_id", executionID)
		// 	}
		// 	logger.Info("queueing ", "execution_id", executionID)
		// 	p.queuedCalls[executionID] = time.Now().String()

		// 	p.executionChannel[executionID] = make(chan interface{})
		// } else if eventName == "event.PipelinePlanned" {
		// 	logger.Info("dequeing ", "execution_id", executionID)
		// 	p.queuedCalls[executionID] = ""
		// }

		// Just do a simple lock for now
		if eventName == "event.PipelinePlan" {
			p.mu.Lock()
			// logger.Info("Before calling h " + eventName + " - " + executionID)
			a, b := h(msg)
			// logger.Info("After calling h " + eventName + " - " + executionID)
			p.mu.Unlock()
			return a, b
		}

		return h(msg)

	}
}
