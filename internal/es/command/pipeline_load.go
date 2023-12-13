package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
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
		slog.Error("invalid command type", "expected", "*event.PipelineLoad", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineLoad")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailedFromPipelineLoad(cmd, err))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromPipelineLoad(cmd, err))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	for _, step := range defn.Steps {
		if step.GetType() == schema.BlockTypePipelineStepContainer || step.GetType() == schema.BlockTypePipelineStepFunction {
			// TODO: If I pass ctx here Docker will initialize OK but then fail when we're trying to use it later. Not sure why, worth investigating
			err := docker.Initialize(context.Background())
			if err != nil {
				slog.Error("Error initializing Docker client", "error", err)
				err2 := h.EventBus.Publish(ctx, event.NewPipelineFailedFromPipelineLoad(cmd, perr.InternalWithMessage("Error initializing Docker client")))
				if err2 != nil {
					slog.Error("Error publishing PipelineFailed event", "error", err2)
				}
				return nil
			}
			break
		}
	}

	e := event.NewPipelineLoadedFromPipelineLoad(cmd, defn)

	return h.EventBus.Publish(ctx, e)
}
