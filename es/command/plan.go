package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PlanHandler CommandHandler

func (h PlanHandler) HandlerName() string {
	return "command.plan"
}

func (h PlanHandler) NewCommand() interface{} {
	return &event.Plan{}
}

func (h PlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.Plan)

	e := event.Planned{
		Event: event.NewFlowEvent(cmd.Event),
	}

	return h.EventBus.Publish(ctx, &e)
}
