package command

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type Plan struct {
	RunID string `json:"run_id"`
}

type PlanHandler CommandHandler

func (h PlanHandler) HandlerName() string {
	return "command.plan"
}

func (h PlanHandler) NewCommand() interface{} {
	return &Plan{}
}

func (h PlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*Plan)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	s, err := state.NewState(cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	e := event.Planned{
		RunID:        cmd.RunID,
		Timestamp:    time.Now(),
		StackID:      xid.New().String(),
		PipelineName: s.PipelineName,
		Input:        s.Input,
	}

	return h.EventBus.Publish(ctx, &e)
}
