package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Finish struct {
	RunID string `json:"run_id"`
}

type FinishHandler CommandHandler

func (h FinishHandler) HandlerName() string {
	return "command.finish"
}

func (h FinishHandler) NewCommand() interface{} {
	return &Finish{}
}

func (h FinishHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*Finish)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), c)

	e := event.Finished{
		RunID:     cmd.RunID,
		Timestamp: time.Now(),
		// TODO - Output
	}

	return h.EventBus.Publish(ctx, &e)
}
