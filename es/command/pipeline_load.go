package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type PipelineLoadHandler CommandHandler

func (h PipelineLoadHandler) HandlerName() string {
	return "command.pipeline_load"
}

func (h PipelineLoadHandler) NewCommand() interface{} {
	return &event.PipelineLoad{}
}

func (h PipelineLoadHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineLoad)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	//defn, err := PipelineDefinition(cmd.Name)
	defn, err := PipelineDefinition("my_pipeline_0")
	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineLoaded{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		Pipeline:  *defn,
	}
	return h.EventBus.Publish(ctx, &e)
}
