package command

import (
	"context"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/util"
	"path"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
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
	logger := fplog.Logger(ctx)
	evt, ok := c.(*event.PipelineCancel)
	if !ok {
		logger.Error("invalid command type", "expected", "*event.PipelineCancel", "actual", c)
		return perr.BadRequestWithMessage("invalid command type expected *event.PipelineCancel")
	}

	eventStoreFilePath := path.Join(util.EventStoreDir(), evt.ExecutionID+".jsonl")
	sanitize.Instance.SanitizeFile(eventStoreFilePath)

	e := event.NewPipelineCanceledFromPipelineCancel(evt)
	return h.EventBus.Publish(ctx, e)
}
