package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Queued EventHandler

func (h Queued) HandlerName() string {
	return "handler.queued"
}

func (Queued) NewEvent() interface{} {
	return &event.Queued{}
}

func (h Queued) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Queued)

	// Next step is to load the mod triggers and pipelines.
	cmd := event.Load{
		Event: event.NewFlowEvent(e.Event),
	}

	return h.CommandBus.Send(ctx, &cmd)
}
