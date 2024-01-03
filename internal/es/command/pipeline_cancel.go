package command

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineCancelHandler CommandHandler

func (h PipelineCancelHandler) HandlerName() string {
	return execution.PipelineCancelCommand.HandlerName()
}

func (h PipelineCancelHandler) NewCommand() interface{} {
	return &event.PipelineCancel{}
}

func (h PipelineCancelHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineCancel)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineCancel", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineCancel")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	e := event.NewPipelineCanceledFromPipelineCancel(cmd)
	return h.EventBus.Publish(ctx, e)
}
