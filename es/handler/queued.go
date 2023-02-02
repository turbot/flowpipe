package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/command"
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

	cmd := &command.Load{
		RunID: e.RunID,
	}

	return h.CommandBus.Send(ctx, cmd)
}
