package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return "command.pipeline_finish"
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelineFinish)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	e := event.PipelineFinished{
		RunID:     cmd.RunID,
		StackID:   cmd.StackID,
		Timestamp: time.Now(),
	}

	return h.EventBus.Publish(ctx, &e)
}
