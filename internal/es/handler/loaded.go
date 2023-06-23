package handler

import (
	"context"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
)

type Loaded EventHandler

func (h Loaded) HandlerName() string {
	return "handler.loaded"
}

func (Loaded) NewEvent() interface{} {
	return &event.Loaded{}
}

func (h Loaded) Handle(ctx context.Context, ei interface{}) error {

	e, ok := ei.(*event.Loaded)
	if !ok {
		fplog.Logger(ctx).Error("invalid event type", "expected", "*event.Loaded", "actual", ei)
		return fperr.BadRequestWithMessage("invalid event type expected *event.Loaded")
	}

	fplog.Logger(ctx).Info("[3] loaded event handler", "executionID", e.Event.ExecutionID)

	// Now that the triggers and pipelines are loaded, we can start mod
	// handling.
	cmd := &event.Start{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, cmd)
}
