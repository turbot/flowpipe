package command

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineStepQueueHandler CommandHandler

func (h PipelineStepQueueHandler) HandlerName() string {
	return "command.pipeline_step_queue"
}

func (h PipelineStepQueueHandler) NewCommand() interface{} {
	return &event.PipelineStepQueue{}
}

// * This is the handler that will actually execute the primitive
func (h PipelineStepQueueHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineStepQueue)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineStepQueue", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineStepQueue")
	}

	logger := fplog.Logger(ctx)
	logger.Info("(10) pipeline_step_queue command handler", "executionID", cmd.Event.ExecutionID, "cmd", cmd)

	logger.Info("Sleeping for delay_ms", "delayMs", cmd.DelayMs)
	time.Sleep(time.Duration(cmd.DelayMs) * time.Millisecond)
	logger.Info("Sleeping for delay_ms complete", "delayMs", cmd.DelayMs)

	e, err := event.NewPipelineStepQueued(event.ForPipelineStepQueue(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineStepQueueToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
