package trigger

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/output"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/fqueue"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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
	ExecuteTrigger() (types.TriggerExecutionResponse, []event.PipelineQueue, error)
	ExecuteTriggerForExecutionID(executionId string, args map[string]interface{}, argsString map[string]string) (types.TriggerExecutionResponse, []event.PipelineQueue, error)
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
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger:    trigger,
				commandBus: commandBus,
				rootMod:    rootMod,
				Fqueue:     fqueue.NewFunctionQueue(trigger.FullName)},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Run() {
	_, _, err := tr.ExecuteTrigger()
	if err != nil {
		slog.Error("Error executing trigger", "trigger", tr.Trigger.Name(), "error", err)
	}
}

func (tr *TriggerRunnerBase) ExecuteTrigger() (types.TriggerExecutionResponse, []event.PipelineQueue, error) {
	return tr.ExecuteTriggerForExecutionID(util.NewExecutionId(), nil, nil)
}

func (tr *TriggerRunnerBase) ExecuteTriggerForExecutionID(executionId string, args map[string]interface{}, argsString map[string]string) (types.TriggerExecutionResponse, []event.PipelineQueue, error) {

	response := types.TriggerExecutionResponse{}
	var triggerRunArgs map[string]interface{}
	if len(args) > 0 || len(argsString) == 0 {
		errs := tr.Trigger.ValidateTriggerParam(args, nil)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return response, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		triggerRunArgs = args
	} else if len(argsString) > 0 {
		coercedArgs, errs := tr.Trigger.CoerceTriggerParams(argsString, nil)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return response, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		triggerRunArgs = coercedArgs
	}

	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return response, nil, perr.BadRequestWithMessage("Pipeline is nil, cannot run trigger")
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	modFullName := tr.Trigger.GetMetadata().ModFullName
	slog.Info("Running trigger", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "mod", modFullName)

	// We can only run trigger from root mod
	canRun := false
	if modFullName != tr.rootMod.FullName {
		for _, m := range tr.rootMod.ResourceMaps.Mods {
			if m.FullName == modFullName {
				canRun = true
				break
			}
		}
	} else {
		canRun = true
	}

	if !canRun {
		slog.Error("Trigger can only be run from root mod and its immediate dependencies", "trigger", tr.Trigger.Name(), "mod", modFullName, "root_mod", tr.rootMod.FullName)
		return response, nil, perr.BadRequestWithMessage("Trigger can only be run from root mod and its immediate dependencies")
	}

	evalContext, err := buildEvalContext(tr.rootMod, tr.Trigger.Params, triggerRunArgs)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return response, nil, perr.InternalWithMessage("Error building eval context")
	}

	latestTrigger, err := db.GetTrigger(tr.Trigger.Name())
	if err != nil {
		slog.Error("Error getting latest trigger", "trigger", tr.Trigger.Name(), "error", err)
		return response, nil, perr.NotFoundWithMessage("trigger not found")
	}

	pipelineArgs, diags := latestTrigger.GetArgs(evalContext)

	if diags.HasErrors() {
		slog.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		err := error_helpers.HclDiagsToError("trigger", diags)
		return response, nil, err
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewEventForExecutionID(executionId),
		PipelineExecutionID: util.NewPipelineExecutionId(),
		Name:                pipelineName,
		Args:                pipelineArgs,
	}

	slog.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if output.IsServerMode {
		output.RenderServerOutput(context.TODO(), types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, tr.Trigger.Name(), pipelineName))
	}

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		if o.IsServerMode {
			o.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error sending pipeline command", err))
		}
		return response, nil, err
	}

	response.Results = map[string]interface{}{}
	response.Results[tr.Trigger.Config.GetType()] = types.PipelineExecutionResponse{
		Flowpipe: types.FlowpipeResponseMetadata{
			ExecutionID:         pipelineCmd.Event.ExecutionID,
			PipelineExecutionID: pipelineCmd.PipelineExecutionID,
			Pipeline:            pipelineCmd.Name,
		},
	}

	response.Flowpipe = types.FlowpipeTriggerResponseMetadata{
		Name: tr.Trigger.FullName,
		Type: tr.Trigger.Config.GetType(),
	}

	return response, []event.PipelineQueue{*pipelineCmd}, nil
}

func (tr *TriggerRunnerBase) GetTrigger() *modconfig.Trigger {
	return tr.Trigger
}

func (tr *TriggerRunnerBase) GetFqueue() *fqueue.FunctionQueue {
	return tr.Fqueue
}

func buildEvalContext(rootMod *modconfig.Mod, triggerParams []modconfig.PipelineParam, triggerRunArgs map[string]interface{}) (*hcl.EvalContext, error) {
	vars := make(map[string]cty.Value)
	if rootMod != nil {
		for _, v := range rootMod.ResourceMaps.Variables {
			vars[v.GetMetadata().ResourceName] = v.Value
		}
	}
	executionVariables := map[string]cty.Value{}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	params := map[string]cty.Value{}

	for _, v := range triggerParams {
		if triggerRunArgs[v.Name] != nil {
			if !v.Type.HasDynamicTypes() {
				val, err := gocty.ToCtyValue(triggerRunArgs[v.Name], v.Type)
				if err != nil {
					return nil, err
				}
				params[v.Name] = val
			} else {
				// we'll do our best here
				val, err := hclhelpers.ConvertInterfaceToCtyValue(triggerRunArgs[v.Name])
				if err != nil {
					return nil, err
				}
				params[v.Name] = val
			}

		} else {
			params[v.Name] = v.Default
		}
	}

	paramsCtyVal := cty.ObjectVal(params)
	executionVariables[schema.BlockTypeParam] = paramsCtyVal

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}

	return evalContext, nil
}
