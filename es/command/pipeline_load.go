package command

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
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

	fmt.Println()
	fmt.Println("in command/pipeline_load.go handle command 1")
	fmt.Println()

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}
	fmt.Println("in command/pipeline_load.go handle command 2")
	fmt.Println()

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}
	fmt.Println("in command/pipeline_load.go handle command 3")
	fmt.Println()

	e, err := event.NewPipelineLoaded(
		event.ForPipelineLoad(cmd),
		event.WithPipelineDefinition(defn))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}
	fmt.Println("in command/pipeline_load.go handle command 3")
	fmt.Println()

	return h.EventBus.Publish(ctx, &e)
}
