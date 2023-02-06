package handler

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	// Now that the triggers and pipelines are loaded, we can start mod
	// handling.
	cmd := &event.Start{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now().UTC(),
	}

	return h.CommandBus.Send(ctx, cmd)
}
