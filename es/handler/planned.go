package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Planned EventHandler

func (h Planned) HandlerName() string {
	return "handler.planned"
}

func (Planned) NewEvent() interface{} {
	return &event.Planned{}
}

func (h Planned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.Planned)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)

	cmd := &event.PipelinePlan{
		RunID:        e.RunID,
		SpanID:       e.SpanID,
		CreatedAt:    time.Now(),
		StackID:      e.StackID,
		PipelineName: e.PipelineName,
		Input:        e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
