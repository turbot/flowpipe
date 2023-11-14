package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineLoadHandler CommandHandler

func (h PipelineLoadHandler) HandlerName() string {
	return execution.PipelineLoadCommand.HandlerName()
}

func (h PipelineLoadHandler) NewCommand() interface{} {
	return &event.PipelineLoad{}
}

// * Path from here:
// * PipelineLoad command handler -> PipelineLoaded event handler -> PipelineStart command
func (h PipelineLoadHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineLoad)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineLoad", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineLoad")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromPipelineLoad(cmd, err))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromPipelineLoad(cmd, err))
	}

	e := event.NewPipelineLoadedFromPipelineLoad(cmd, defn)

	return h.EventBus.Publish(ctx, e)
}
