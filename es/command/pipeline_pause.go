package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
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

	evt, ok := c.(*event.PipelinePause)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelinePause", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelinePause")
	}

	logger.Info("(7) pipeline_pause command handler")

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePauseToPipelineFailed(evt, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[evt.PipelineExecutionID]
	if pe == nil {
		logger.Error("Can't pause pipeline execution that doesn't exist", "pipeline_execution_id", evt.PipelineExecutionID)
		return nil
	}

	if pe.Status != "started" && pe.Status != "queued" {
		logger.Error("Can't pause pipeline execution that is not started or queued", "pipeline_execution_id", evt.PipelineExecutionID, "pipelineStatus", pe.Status)

		// TODO: This is a bit of a hack. We should be able to return an error here, but returning an error is causing
		// TODO: Watermill retry the message ... infinitely. There's something wrong here, the example clearly shows
		// TODO: that command can and should return an error if there's a problem.
		// return fperr.BadRequestWithMessage("Can't pause pipeline execution that is not started or queued")
		return nil
	}

	e, err := event.NewPipelinePaused(event.ForPipelinePause(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelinePauseToPipelineFailed(evt, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
