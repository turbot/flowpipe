package trigger

import (
	"context"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerBase struct {
	ctx       context.Context
	trigger   pipeline.ITrigger
	esService *es.ESService
}

type ITriggerRunner interface {
	Run()
}

func NewTriggerRunner(ctx context.Context, esService *es.ESService, trigger pipeline.ITrigger) ITriggerRunner {
	switch trigger.(type) {
	case *pipeline.TriggerSchedule, *pipeline.TriggerInterval:
		return &TriggerRunnerBase{
			ctx:       ctx,
			trigger:   trigger,
			esService: esService,
		}
	case *pipeline.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				ctx:       ctx,
				trigger:   trigger,
				esService: esService,
			},
		}
	case *pipeline.TriggerHttp:
		return &TriggerRunnerHttp{
			TriggerRunnerBase: TriggerRunnerBase{
				ctx:       ctx,
				trigger:   trigger,
				esService: esService,
			},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Run() {
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
		Args:                tr.trigger.GetArgs(),
	}

	logger.Info("Trigger fired", "trigger", tr.trigger.GetName(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if err := tr.esService.Send(pipelineCmd); err != nil {
		logger.Error("Error sending pipeline command", "error", err)
		return
	}
}

type TriggerRunnerQuery struct {
	TriggerRunnerBase
}

func (tr *TriggerRunnerQuery) Run() {
}

type TriggerRunnerHttp struct {
	TriggerRunnerBase
}

func (tr *TriggerRunnerHttp) Run() {
}
