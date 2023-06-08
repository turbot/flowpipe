package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelinePauseHandler CommandHandler

func (h PipelinePauseHandler) HandlerName() string {
	return "command.pipeline_pause"
}

func (h PipelinePauseHandler) NewCommand() interface{} {
	return &event.PipelinePause{}
}

// pipeline_pause command handler
// issue this to pause a pipeline execution
func (h PipelinePauseHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	cmd, ok := c.(*event.PipelinePause)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelinePause", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelinePause")
	}

	logger.Info("(7) pipeline_pause command handler")

	e, err := event.NewPipelinePaused(event.ForPipelinePause(cmd))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePauseToPipelineFailed(cmd, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
