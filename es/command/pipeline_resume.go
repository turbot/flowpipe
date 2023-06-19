package command

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/es/execution"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

type PipelineResumeHandler CommandHandler

func (h PipelineResumeHandler) HandlerName() string {
	return "command.pipeline_resume"
}

func (h PipelineResumeHandler) NewCommand() interface{} {
	return &event.PipelineResume{}
}

// pipeline_resume command handler
// issue this to pause a pipeline execution
func (h PipelineResumeHandler) Handle(ctx context.Context, c interface{}) error {
	logger := fplog.Logger(ctx)

	evt, ok := c.(*event.PipelineResume)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineResume", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineResume")
	}

	logger.Info("(9) pipeline_resume command handler")

	ex, err := execution.NewExecution(ctx, execution.WithEvent(evt.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineResumeToPipelineFailed(evt, err)))
	}

	// Convenience
	pe := ex.PipelineExecutions[evt.PipelineExecutionID]
	if pe == nil {
		logger.Error("Can't resume pipeline execution that doesn't exist", "pipeline_execution_id", evt.PipelineExecutionID)
		return fperr.BadRequestWithMessage("Can't resume pipeline execution that doesn't exist")
	}

	if !pe.IsPaused() {
		logger.Error("Can't resume pipeline execution that is not paused", "pipeline_execution_id", evt.PipelineExecutionID, "pipelineStatus", pe.Status)
		return fperr.BadRequestWithMessage("Can't resume pipeline execution that is not paused")
	}

	e, err := event.NewPipelineResumed(event.ForPipelineResume(evt))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineResumeToPipelineFailed(evt, err)))
	}
	return h.EventBus.Publish(ctx, &e)
}
