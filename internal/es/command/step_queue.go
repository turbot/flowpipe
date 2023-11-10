package command

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueueHandler CommandHandler

func (h StepQueueHandler) HandlerName() string {
	return execution.StepQueueCommand.HandlerName()
}

func (h StepQueueHandler) NewCommand() interface{} {
	return &event.StepQueue{}
}

// * This is the handler that will actually execute the primitive
func (h StepQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.StepQueue)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.StepQueue", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.StepQueue")
	}

	logger := fplog.Logger(ctx)

	if cmd.DelayMs > 0 {
		logger.Info("Sleeping for delay_ms", "delayMs", cmd.DelayMs)
		time.Sleep(time.Duration(cmd.DelayMs) * time.Millisecond)
		logger.Info("Sleeping for delay_ms complete", "delayMs", cmd.DelayMs)
	}

	e, err := event.NewStepQueued(event.ForStepQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForStepQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, e)
}
