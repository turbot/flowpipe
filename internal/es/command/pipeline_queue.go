package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineQueueHandler CommandHandler

func (h PipelineQueueHandler) HandlerName() string {
	return execution.PipelineQueueCommand.HandlerName()
}

func (h PipelineQueueHandler) NewCommand() interface{} {
	return &event.PipelineQueue{}
}

func (h PipelineQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineQueue)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineQueue")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e, err := event.NewPipelineQueued(event.ForPipelineQueue(cmd))
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineQueueToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("error publishing event", "error", err2)
		}
		return nil
	}

	err = store.StartPipeline(cmd.Event.ExecutionID, cmd.Name)
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineQueueToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("error publishing event", "error", err2)
		}
		return nil
	}

	return h.EventBus.Publish(ctx, e)
}
