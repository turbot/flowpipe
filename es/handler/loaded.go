package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Loaded EventHandler

func (h Loaded) HandlerName() string {
	return "handler.loaded"
}

func (Loaded) NewEvent() interface{} {
	return &event.Loaded{}
}

func (h Loaded) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Loaded)

	// Now that the triggers and pipelines are loaded, we can start mod
	// handling.
	cmd := &event.Start{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, cmd)
}
