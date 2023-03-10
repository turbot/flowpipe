package command

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
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

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		e := event.PipelineFailed{
			Event:        event.NewFlowEvent(cmd.Event),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		e := event.PipelineFailed{
			Event:        event.NewFlowEvent(cmd.Event),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e, err := event.NewPipelineLoaded(
		event.ForPipelineLoad(cmd),
		event.WithPipelineDefinition(defn))
	if err != nil {
		return err
	}

	return h.EventBus.Publish(ctx, &e)
}
