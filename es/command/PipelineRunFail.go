package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineRunFail struct {
	RunID        string `json:"run_id"`
	ErrorMessage string `json:"error_message"`
}

type PipelineRunFailHandler CommandHandler

func (h PipelineRunFailHandler) HandlerName() string {
	return "pipeline.run.fail"
}

func (h PipelineRunFailHandler) NewCommand() interface{} {
	return &PipelineRunFail{}
}

func (h PipelineRunFailHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunFail)

	e := event.PipelineRunFailed{
		RunID:        cmd.RunID,
		Timestamp:    time.Now(),
		ErrorMessage: cmd.ErrorMessage,
	}

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), e)

	return h.EventBus.Publish(ctx, &e)
}
