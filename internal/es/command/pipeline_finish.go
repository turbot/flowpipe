package command

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return "command.pipeline_finish"
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd, ok := c.(*event.PipelineFinish)
	if !ok {
		fplog.Logger(ctx).Error("invalid command type", "expected", "*event.PipelineFinish", "actual", c)
		return fperr.BadRequestWithMessage("invalid command type expected *event.PipelineFinish")
	}

	fplog.Logger(ctx).Info("(5) pipeline_finish command handler")

	var output types.StepOutput
	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	if defn.Output != nil {

		// Parse the input template once
		t, err := template.New("output").Parse(*defn.Output)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

		data, err := ex.PipelineStepOutputs(cmd.PipelineExecutionID)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

		var outputBuffer bytes.Buffer
		err = t.Execute(&outputBuffer, data)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}
		err = json.Unmarshal(outputBuffer.Bytes(), &output)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

	}

	e, err := event.NewPipelineFinished(event.ForPipelineFinish(cmd, &output))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(ctx, event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	return h.EventBus.Publish(ctx, &e)
}
