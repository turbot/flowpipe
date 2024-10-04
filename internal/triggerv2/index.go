package triggerv2

import (
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/db"
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
	Trigger *modconfig.Trigger
	rootMod *modconfig.Mod
}

type TriggerRunner interface {
	Validate(args map[string]interface{}, argsString map[string]string) (map[string]interface{}, error)
	GetPipelineArgs(triggerRunArgs map[string]interface{}) (modconfig.Input, error)
}

func NewTriggerRunner(trigger *modconfig.Trigger) TriggerRunner {

	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule:
		return &TriggerRunnerBase{
			Trigger: trigger,
			rootMod: trigger.GetMod(),
		}
	case *modconfig.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger: trigger,
				rootMod: trigger.GetMod(),
			},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Validate(args map[string]interface{}, argsString map[string]string) (map[string]interface{}, error) {
	var triggerRunArgs map[string]interface{}

	evalContext, err := buildEvalContext(tr.rootMod, tr.Trigger.Params, nil)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return nil, perr.InternalWithMessage("Error building eval context")
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

func (tr *TriggerRunnerBase) GetPipelineArgs(triggerRunArgs map[string]interface{}) (modconfig.Input, error) {

	evalContext, err := buildEvalContext(tr.rootMod, tr.Trigger.Params, triggerRunArgs)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return nil, perr.InternalWithMessage("Error building eval context")
	}

	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return nil, perr.BadRequestWithMessage("Pipeline is nil, cannot run trigger")
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
		return nil, perr.BadRequestWithMessage("Trigger can only be run from root mod and its immediate dependencies")
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
