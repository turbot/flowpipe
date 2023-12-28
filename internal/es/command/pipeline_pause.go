package command

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelinePauseHandler CommandHandler

func (h PipelinePauseHandler) HandlerName() string {
	return execution.PipelinePauseCommand.HandlerName()
}

func (h PipelinePauseHandler) NewCommand() interface{} {
	return &event.PipelinePause{}
}

// pipeline_pause command handler
// issue this to pause a pipeline execution
func (h PipelinePauseHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelinePause)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelinePause", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelinePause")
	}

	ex, err := execution.GetExecution(cmd.Event.ExecutionID)
	if err != nil {
		slog.Error("pipeline_pause: Error loading pipeline execution", "error", err)
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePauseToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}

	pex := ex.PipelineExecutions[cmd.PipelineExecutionID]

	if pex.Status != "started" && pex.Status != "queued" {
		slog.Error("Can't pause pipeline execution that is not started or queued", "pipeline_execution_id", cmd.PipelineExecutionID, "pipelineStatus", pex.Status)
		return perr.BadRequestWithMessage("Can't pause pipeline execution that is not started or queued")
	}

	e, err := event.NewPipelinePaused(event.ForPipelinePause(cmd))
	if err != nil {
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelinePauseToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}
	return h.EventBus.Publish(ctx, e)
}
