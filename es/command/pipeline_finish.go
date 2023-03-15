package command

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineFinishHandler CommandHandler

func (h PipelineFinishHandler) HandlerName() string {
	return "command.pipeline_finish"
}

func (h PipelineFinishHandler) NewCommand() interface{} {
	return &event.PipelineFinish{}
}

func (h PipelineFinishHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineFinish)

	var output pipeline.PipelineOutput

	ex, err := execution.NewExecution(ctx, execution.WithEvent(cmd.Event))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	defn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	if defn.Output != "" {

		// Parse the input template once
		t, err := template.New("output").Parse(defn.Output)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

		stepOutputs, err := ex.PipelineStepOutputs(cmd.PipelineExecutionID)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

		var outputBuffer bytes.Buffer
		err = t.Execute(&outputBuffer, stepOutputs)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}
		err = json.Unmarshal(outputBuffer.Bytes(), &output)
		if err != nil {
			return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
		}

	}

	e, err := event.NewPipelineFinished(event.ForPipelineFinish(cmd, &output))
	if err != nil {
		return h.EventBus.Publish(ctx, event.NewPipelineFailed(event.ForPipelineFinishToPipelineFailed(cmd, err)))
	}

	return h.EventBus.Publish(ctx, &e)
}
