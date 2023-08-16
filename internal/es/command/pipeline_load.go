package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type PipelineLoadHandler CommandHandler

func (h PipelineLoadHandler) HandlerName() string {
	return "command.pipeline_load"
}

func (h PipelineLoadHandler) NewCommand() interface{} {
	return &event.PipelineLoad{}
}

// Path from here:
// * PipelineLoad command handler -> PipelineLoaded event handler -> PipelineStart command
func (h PipelineLoadHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineLoad)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineLoad", "actual", c)
		return pcerr.BadRequestWithMessage("invalid command type expected *event.PipelineLoad")
	}

	// ? new execution here? is it because I'm finally running the pipeline?
	// ? doesn't look like the execution is used for anything else apart from loading a pipeline definition
	// ? and we need the execution "instance" so we can get the pipeline name from the pipeline execution id
	// ? should we have a main store for this rather than creating a new execution instance?
	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}

	e, err := event.NewPipelineLoaded(
		event.ForPipelineLoad(cmd),
		event.WithPipelineDefinition(defn))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineLoadToPipelineFailed(cmd, err)))
	}

	return h.EventBus.Publish(ctx, &e)
}
