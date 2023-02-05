package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
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

	s, err := state.NewState(cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := PipelineDefinition(s.PipelineName)
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
