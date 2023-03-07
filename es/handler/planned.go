package handler

import (
	"context"

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

	cmd := &event.PipelinePlan{
		Event:        event.NewFlowEvent(e.Event),
		PipelineName: e.PipelineName,
		Input:        e.Input,
	}

	return h.CommandBus.Send(ctx, cmd)
}
