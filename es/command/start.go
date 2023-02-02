package command

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type Start struct {
	RunID string `json:"run_id"`
}

type StartHandler CommandHandler

func (h StartHandler) HandlerName() string {
	return "command.start"
}

func (h StartHandler) NewCommand() interface{} {
	return &Start{}
}

func (h StartHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*Start)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	s, err := state.NewState(cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	e := event.Started{
		RunID:        cmd.RunID,
		Timestamp:    time.Now(),
		StackID:      xid.New().String(),
		PipelineName: s.PipelineName,
		Input:        s.Input,
	}

	return h.EventBus.Publish(ctx, &e)
}
