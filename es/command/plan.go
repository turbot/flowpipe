package command

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
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

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	/*
		s, err := state.NewState(cmd.SpanID)
		if err != nil {
			// TODO - should this return a failed event? how are errors caught here?
			return err
		}
	*/

	e := event.Planned{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now().UTC(),
		StackID:   xid.New().String(),
	}

	return h.EventBus.Publish(ctx, &e)
}
