package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type Queued EventHandler

func (h Queued) HandlerName() string {
	return "handler.queued"
}

func (Queued) NewEvent() interface{} {
	return &event.Queued{}
}

func (h Queued) Handle(ctx context.Context, ei interface{}) error {

	e, ok := ei.(*event.Queued)
	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.Queued", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.Queued")
	}

	fplog.Logger(ctx).Info("[15] queued event handler", "executionID", e.Event.ExecutionID)

	// Next step is to load the mod triggers and pipelines.
	cmd := event.Load{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, &cmd)
}
