package trigger

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerBase struct {
	Ctx       context.Context
	Trigger   *modconfig.Trigger
	EsService *es.ESService
}

type ITriggerRunner interface {
	Run()
}

func NewTriggerRunner(ctx context.Context, esService *es.ESService, trigger *modconfig.Trigger) ITriggerRunner {
	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule, *modconfig.TriggerInterval:
		return &TriggerRunnerBase{
			Ctx:       ctx,
			Trigger:   trigger,
			EsService: esService,
		}
	case *modconfig.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Ctx:       ctx,
				Trigger:   trigger,
				EsService: esService,
			},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Run() {
	logger := fplog.Logger(tr.Ctx)

	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		logger.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	piplineArgs, diags := tr.Trigger.GetArgs(nil)

	if diags.HasErrors() {
		logger.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(tr.Ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                piplineArgs,
	}

	logger.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if err := tr.EsService.Send(pipelineCmd); err != nil {
		logger.Error("Error sending pipeline command", "error", err)
		return
	}
}

type TriggerRunnerQuery struct {
	TriggerRunnerBase
}

func (tr *TriggerRunnerQuery) Run() {
}
