package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/perr"
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
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelinePause")
	}

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePauseToPipelineFailed(evt, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[evt.PipelineExecutionID]
	if pe == nil {
		logger.Error("Can't pause pipeline execution that doesn't exist", "pipeline_execution_id", evt.PipelineExecutionID)
		return perr.BadRequestWithMessage("Can't pause pipeline execution that doesn't exist")
	}

	if pe.Status != "started" && pe.Status != "queued" {
		logger.Error("Can't pause pipeline execution that is not started or queued", "pipeline_execution_id", evt.PipelineExecutionID, "pipelineStatus", pe.Status)
		return perr.BadRequestWithMessage("Can't pause pipeline execution that is not started or queued")
	}

	e, err := event.NewPipelinePaused(event.ForPipelinePause(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePauseToPipelineFailed(evt, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
