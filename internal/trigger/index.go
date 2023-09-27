package trigger

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/funcs"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
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

	modFullName := tr.Trigger.GetMetadata().ModFullName
	logger.Info("Running trigger", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "mod", modFullName)

	// We can only run trigger from root mod

	mod := tr.EsService.RootMod

	if modFullName != mod.FullName {
		logger.Error("Trigger can only be run from root mod", "trigger", tr.Trigger.Name(), "mod", modFullName, "root_mod", mod.FullName)
		return
	}

	vars := map[string]cty.Value{}
	for _, v := range mod.ResourceMaps.Variables {
		vars[v.GetMetadata().ResourceName] = v.Value
	}

	executionVariables := map[string]cty.Value{}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString("work.dir")),
	}

	pipelineArgs, diags := tr.Trigger.GetArgs(evalContext)

	if diags.HasErrors() {
		logger.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(tr.Ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                pipelineArgs,
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
