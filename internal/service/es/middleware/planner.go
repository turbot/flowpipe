package middleware

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/turbot/flowpipe/internal/fplog"
)

type PlannerControl struct {
	Ctx context.Context
}

func (p PlannerControl) Middleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := fplog.Logger(p.Ctx)

		// eventName := msg.Metadata.Get("name")
		// if eventName != "event.PipelinePlan" && eventName != "event.PipelinePlanned" && eventName != "event.PipelineFinished" && eventName != "event.PipelineCanceled" && eventName != "event.PipelineFailed" && eventName != "event.PipelineFinished" {
		// 	return h(msg)
		// }

		// logger.Info("***** Planner control middleware", "msg", msg)

		return h(msg)
	}
}
