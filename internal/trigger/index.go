package trigger

import (
	"context"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"

	_ "github.com/mattn/go-sqlite3"
)

type TriggerRunnerBase struct {
	Trigger    *modconfig.Trigger
	commandBus handler.FpCommandBus
	rootMod    *modconfig.Mod
}

type TriggerRunner interface {
	Run()
}

func NewTriggerRunner(ctx context.Context, commandBus handler.FpCommandBus, rootMod *modconfig.Mod, trigger *modconfig.Trigger) TriggerRunner {
	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule, *modconfig.TriggerInterval:
		return &TriggerRunnerBase{
			Trigger:    trigger,
			commandBus: commandBus,
			rootMod:    rootMod,
		}
	case *modconfig.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger:    trigger,
				commandBus: commandBus,
				rootMod:    rootMod},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Run() {
	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	modFullName := tr.Trigger.GetMetadata().ModFullName
	slog.Info("Running trigger", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "mod", modFullName)

	// We can only run trigger from root mod

	if modFullName != tr.rootMod.FullName {
		slog.Error("Trigger can only be run from root mod", "trigger", tr.Trigger.Name(), "mod", modFullName, "root_mod", tr.rootMod.FullName)
		return
	}

	vars := map[string]cty.Value{}
	for _, v := range tr.rootMod.ResourceMaps.Variables {
		vars[v.GetMetadata().ResourceName] = v.Value
	}

	executionVariables := map[string]cty.Value{}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	evalContext, err := buildEvalContext(tr.rootMod)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return
	}

	pipelineArgs, diags := tr.Trigger.GetArgs(evalContext)

	if diags.HasErrors() {
		slog.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                pipelineArgs,
	}

	slog.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		return
	}
}

func buildEvalContext(rootMod *modconfig.Mod) (*hcl.EvalContext, error) {
	vars := make(map[string]cty.Value)
	if rootMod != nil {
		for _, v := range rootMod.ResourceMaps.Variables {
			vars[v.GetMetadata().ResourceName] = v.Value
		}
	}

	executionVariables := map[string]cty.Value{}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}
	return evalContext, nil
}
