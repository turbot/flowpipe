package command

import (
	"context"

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

	s, err := state.NewState(ctx, cmd.Event)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		e := event.PipelineFailed{
			Event:        event.NewFlowEvent(cmd.Event),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineLoaded{
		Event:    cmd.Event,
		Pipeline: *defn,
	}
	return h.EventBus.Publish(ctx, &e)
}
