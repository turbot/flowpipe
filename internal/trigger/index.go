package trigger

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type TriggerRunnerBase struct {
	ExecutionID        string
	TriggerExecutionID string
	Trigger            *modconfig.Trigger
	rootMod            modconfig.ModI
	Type               string
}

type TriggerRunner interface {
	GetTrigger() *modconfig.Trigger
	GetPipelineQueuesWithArgs(ctx context.Context, args map[string]interface{}, argsString map[string]string) ([]*event.PipelineQueue, error)
	GetTriggerResponse([]*event.PipelineQueue) (types.TriggerExecutionResponse, error)
}

func NewTriggerRunner(trigger *modconfig.Trigger, executionID, triggerExecutionID string) TriggerRunner {

	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule:
		return &TriggerRunnerBase{
			Trigger:            trigger,
			rootMod:            trigger.GetMod(),
			ExecutionID:        executionID,
			TriggerExecutionID: triggerExecutionID,
			Type:               "schedule",
		}
	case *modconfig.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger:            trigger,
				rootMod:            trigger.GetMod(),
				ExecutionID:        executionID,
				TriggerExecutionID: triggerExecutionID,
				Type:               "query",
			},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) GetTrigger() *modconfig.Trigger {
	return tr.Trigger
}

func (tr *TriggerRunnerBase) GetPipelineQueuesWithArgs(ctx context.Context, args map[string]interface{}, argsString map[string]string) ([]*event.PipelineQueue, error) {
	triggerRunArgs, err := tr.validate(args, argsString)

	if err != nil {
		slog.Error("Error validating trigger", "error", err)
		return nil, err
	}

	triggerArgs, err := tr.getTriggerArgs(triggerRunArgs)
	if err != nil {
		return nil, err
	}

	cmds, err := tr.execute(ctx, tr.ExecutionID, triggerArgs, tr.Trigger)
	if err != nil {
		slog.Error("Error sending pipeline command", "error", err)

		return nil, err
	}

	return cmds, nil

}

func (tr *TriggerRunnerBase) GetTriggerResponse(pipelineCmds []*event.PipelineQueue) (types.TriggerExecutionResponse, error) {
	response := types.TriggerExecutionResponse{}

	if len(pipelineCmds) == 0 {
		return response, perr.NotFoundWithMessage("no pipeline commands found")
	}

	if len(pipelineCmds) > 1 {
		return response, perr.BadRequestWithMessage("multiple pipeline commands found")
	}

	pipelineCmd := pipelineCmds[0]

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

	return response, nil
}

func (tr *TriggerRunnerBase) validate(args map[string]interface{}, argsString map[string]string) (map[string]interface{}, error) {
	var triggerRunArgs map[string]interface{}

	evalContext, err := buildEvalContextForTriggerExecution(tr.rootMod, tr.Trigger.Params, tr.Trigger.Config, nil)
	if err != nil {
		slog.Error("Error building eval context (trigger validation)", "error", err)
		return nil, perr.InternalWithMessage("Error building eval context: " + err.Error())
	}

	if len(args) > 0 || len(argsString) == 0 {
		errs := parse.ValidateParams(tr.Trigger, args, evalContext)

		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		triggerRunArgs = args
	} else if len(argsString) > 0 {
		coercedArgs, errs := parse.CoerceParams(tr.Trigger, argsString, evalContext)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		triggerRunArgs = coercedArgs
	}

	return triggerRunArgs, nil
}

func (tr *TriggerRunnerBase) getTriggerArgs(triggerRunArgs map[string]interface{}) (modconfig.Input, error) {

	evalContext, err := buildEvalContextForTriggerExecution(tr.rootMod, tr.Trigger.Params, tr.Trigger.Config, triggerRunArgs)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return nil, perr.InternalWithMessage("Error building eval context")
	}

	latestTrigger, err := db.GetTrigger(tr.Trigger.Name())
	if err != nil {
		slog.Error("Error getting latest trigger", "trigger", tr.Trigger.Name(), "error", err)
		return nil, perr.NotFoundWithMessage("trigger not found")
	}

	pipelineArgs, diags := latestTrigger.GetArgs(evalContext)

	if diags.HasErrors() {
		slog.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		err := error_helpers.HclDiagsToError("trigger", diags)
		return nil, err
	}

	return pipelineArgs, nil
}

func (tr *TriggerRunnerBase) execute(ctx context.Context, executionID string, triggerArgs modconfig.Input, trg *modconfig.Trigger) ([]*event.PipelineQueue, error) {
	pipelineDefn := trg.Pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewEventForExecutionID(executionID),
		PipelineExecutionID: util.NewPipelineExecutionId(),
		Name:                pipelineName,
		Args:                triggerArgs,
		Trigger:             trg.Name(),
	}

	slog.Info("Trigger fired", "trigger", trg.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if output.IsServerMode {
		output.RenderServerOutput(ctx, types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, trg.Name(), pipelineName))
	}

	return []*event.PipelineQueue{pipelineCmd}, nil
}

func buildEvalContextForTriggerExecution(rootMod modconfig.ModI, defnTriggerParams []modconfig.PipelineParam, triggerConfig modconfig.TriggerConfig, triggerRunArgs map[string]interface{}) (*hcl.EvalContext, error) {

	executionVariables := map[string]cty.Value{}

	// populate the variables and locals
	// build a variables map _excluding_ late binding vars, and a separate map for late binding vars
	var modVars map[string]*modconfig.Variable
	if rootMod != nil {
		modVars = rootMod.ResourceMaps.Variables
	}
	variablesMap, _, lateBindingVarDeps := parse.VariableValueCtyMap(modVars, true)

	// add these to eval context
	executionVariables[constants.LateBindingVarsKey] = cty.ObjectVal(lateBindingVarDeps)
	for _, variable := range modVars {
		variablesMap[variable.ShortName] = variable.Value
	}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(variablesMap)

	localsMap := make(map[string]cty.Value)
	if rootMod != nil {
		for _, local := range rootMod.ResourceMaps.Locals {
			localsMap[local.ShortName] = local.Value
		}
		executionVariables[schema.AttributeLocal] = cty.ObjectVal(localsMap)
	}

	runParams := map[string]cty.Value{}

	for _, v := range defnTriggerParams {
		if triggerRunArgs[v.Name] != nil {
			if !v.Type.HasDynamicTypes() {
				val, err := gocty.ToCtyValue(triggerRunArgs[v.Name], v.Type)
				if err != nil {
					return nil, err
				}
				runParams[v.Name] = val
			} else {
				// we'll do our best here
				val, err := hclhelpers.ConvertInterfaceToCtyValue(triggerRunArgs[v.Name])
				if err != nil {
					return nil, err
				}
				runParams[v.Name] = val
			}

		} else {
			runParams[v.Name] = v.Default
		}
	}

	paramsCtyVal := cty.ObjectVal(runParams)
	executionVariables[schema.BlockTypeParam] = paramsCtyVal

	// add in mod connection dependendencies
	allConnectionsDependsOn := triggerConfig.GetConnectionDependsOn()
	if rootMod != nil {
		allConnectionsDependsOn = append(allConnectionsDependsOn, rootMod.GetConnectionDependsOn()...)
	}
	connectionMap, paramsMap, varMap, err := execution.BuildConnectionMapForEvalContext(allConnectionsDependsOn, runParams, variablesMap, defnTriggerParams)
	if err != nil {
		return nil, err
	}

	executionVariables[schema.BlockTypeConnection] = cty.ObjectVal(connectionMap)
	executionVariables[schema.BlockTypeParam] = cty.ObjectVal(paramsMap)
	executionVariables[schema.AttributeVar] = cty.ObjectVal(varMap)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}

	return evalContext, nil
}
