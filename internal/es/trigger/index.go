package trigger

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunner struct {
	ctx       context.Context
	trigger   types.ITrigger
	esService *es.ESService
}

func NewTriggerRunner(ctx context.Context, esService *es.ESService, trigger types.ITrigger) *TriggerRunner {
	return &TriggerRunner{
		ctx:       ctx,
		trigger:   trigger,
		esService: esService,
	}
}

func (tr *TriggerRunner) Run() {
	logger := fplog.Logger(tr.ctx)

	pipeline := tr.trigger.GetPipeline()

	if pipeline == cty.NilVal {
		logger.Error("Pipeline is nil, cannot run trigger", "trigger", tr.trigger.GetName())
		return
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(tr.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
	}

	logger.Info("Trigger fired", "trigger", tr.trigger.GetName(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if err := tr.esService.Send(pipelineCmd); err != nil {
		logger.Error("Error sending pipeline command", "error", err)
		return
	}
}
