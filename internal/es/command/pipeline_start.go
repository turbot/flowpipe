package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineStartHandler CommandHandler

func (h PipelineStartHandler) HandlerName() string {
	return execution.PipelineStartCommand.HandlerName()
}

func (h PipelineStartHandler) NewCommand() interface{} {
	return &event.PipelineStart{}
}

func (h PipelineStartHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStart)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineStart", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineStart")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e, err := event.NewPipelineStarted(event.ForPipelineStart(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStartToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
