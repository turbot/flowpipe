package trigger

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/fqueue"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerBase struct {
	Trigger    *modconfig.Trigger
	commandBus handler.FpCommandBus
	rootMod    *modconfig.Mod
	Fqueue     *fqueue.FunctionQueue
}

type TriggerRunner interface {
	Run()
	GetTrigger() *modconfig.Trigger
	GetFqueue() *fqueue.FunctionQueue
}

func NewTriggerRunner(ctx context.Context, commandBus handler.FpCommandBus, rootMod *modconfig.Mod, trigger *modconfig.Trigger) TriggerRunner {

	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule:
		return &TriggerRunnerBase{
			Trigger:    trigger,
			commandBus: commandBus,
			rootMod:    rootMod,
			Fqueue:     fqueue.NewFunctionQueue(trigger.FullName),
		}
	case *modconfig.TriggerQuery:
		dbDir := filepaths.ModDir()
		dbFile := filepath.Join(dbDir, "flowpipe.db")

		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger:    trigger,
				commandBus: commandBus,
				rootMod:    rootMod,
				Fqueue:     fqueue.NewFunctionQueue(trigger.FullName)},
			DatabasePath: dbFile,
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

	if output.IsServerMode {
		output.RenderServerOutput(context.TODO(), types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, tr.Trigger.Name(), pipelineName))
	}

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		return
	}
}

func (tr *TriggerRunnerBase) GetTrigger() *modconfig.Trigger {
	return tr.Trigger
}

func (tr *TriggerRunnerBase) GetFqueue() *fqueue.FunctionQueue {
	return tr.Fqueue
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
