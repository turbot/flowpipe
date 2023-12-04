package command

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/perr"
	"log/slog"
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
	evt, ok := c.(*event.PipelineResume)
	if !ok {
		slog.Error("invalid command type", "expected", "*event.PipelineResume", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineResume")
	}

	slog.Info("(9) pipeline_resume command handler")

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineResumeToPipelineFailed(evt, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[evt.PipelineExecutionID]
	if pe == nil {
		slog.Error("Can't resume pipeline execution that doesn't exist", "pipeline_execution_id", evt.PipelineExecutionID)
		return perr.BadRequestWithMessage("Can't resume pipeline execution that doesn't exist")
	}

	if !pe.IsPaused() {
		slog.Error("Can't resume pipeline execution that is not paused", "pipeline_execution_id", evt.PipelineExecutionID, "pipelineStatus", pe.Status)
		return perr.BadRequestWithMessage("Can't resume pipeline execution that is not paused")
	}

	e := event.NewPipelineResumedFromPipelineResume(evt)

	return h.EventBus.Publish(ctx, e)
}
