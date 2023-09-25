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

type TriggerRunnerHttp struct {
	TriggerRunnerBase
}

func (tr *TriggerRunnerHttp) Run(c context.Context, data map[string]interface{}) error {
	// executionVariables := map[string]cty.Value{}

	// selfObject := map[string]cty.Value{}
	// for k, v := range data {
	// 	ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(v)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	selfObject[k] = ctyVal
	// }

	// executionVariables["self"] = cty.ObjectVal(selfObject)

	// evalContext := &hcl.EvalContext{
	// 	Variables: executionVariables,
	// 	Functions: funcs.ContextFunctions(viper.GetString("work.dir")),
	// }

	// pipelineArgs, diags := tr.Trigger.GetArgs(evalContext)
	// if diags.HasErrors() {
	// 	return error_helpers.HclDiagsToError("trigger", diags)
	// }

	// pipeline := tr.Trigger.GetPipeline()
	// pipelineName := pipeline.AsValueMap()["name"].AsString()

	// pipelineCmd := &event.PipelineQueue{
	// 	Event:               event.NewExecutionEvent(c),
	// 	PipelineExecutionID: util.NewPipelineExecutionID(),
	// 	Name:                pipelineName,
	// }

	// pipelineCmd.Args = pipelineArgs

	// if err := api.EsService.Send(pipelineCmd); err != nil {
	// 	return err
	// }

	// response := types.RunPipelineResponse{
	// 	ExecutionID:           pipelineCmd.Event.ExecutionID,
	// 	PipelineExecutionID:   pipelineCmd.PipelineExecutionID,
	// 	ParentStepExecutionID: pipelineCmd.ParentStepExecutionID,
	// }
	return nil

}
