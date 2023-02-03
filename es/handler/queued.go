package handler

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	// Next step is to load the mod triggers and pipelines.
	cmd := event.Load{
		RunID:     e.RunID,
		SpanID:    e.SpanID,
		CreatedAt: time.Now(),
	}

	return h.CommandBus.Send(ctx, &cmd)
}
