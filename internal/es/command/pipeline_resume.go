package command

import (
	"context"

	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
)

type PipelineResumeHandler CommandHandler

func (h PipelineResumeHandler) HandlerName() string {
	return execution.PipelineResumeCommand.HandlerName()
}

func (h PipelineResumeHandler) NewCommand() interface{} {
	return &event.PipelineResume{}
}

// pipeline_resume command handler
// issue this to pause a pipeline execution
func (h PipelineResumeHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineResume)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineResume", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineResume")
	}

	plannerMutex := event.GetEventStoreMutex(cmd.Event.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		plannerMutex.Unlock()
	}()

	ex, err := execution.GetExecution(cmd.Event.ExecutionID)
	if err != nil {
		slog.Error("pipeline_resume: Error loading pipeline execution", "error", err)
		err2 := h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineResumeToPipelineFailed(cmd, err)))
		if err2 != nil {
			slog.Error("Error publishing PipelineFailed event", "error", err2)
		}
		return nil
	}
	pe := ex.PipelineExecutions[cmd.PipelineExecutionID]

	if !pe.IsPaused() {
		slog.Error("Can't resume pipeline execution that is not paused", "pipeline_execution_id", cmd.PipelineExecutionID, "pipelineStatus", pe.Status)
		return perr.BadRequestWithMessage("Can't resume pipeline execution that is not paused")
	}

	e := event.NewPipelineResumedFromPipelineResume(cmd)

	return h.EventBus.Publish(ctx, e)
}
