package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Execution represents the current state of an execution. A single execution
// is tied to a trigger (webhook, cronjob, etc) and may result in multiple
// pipelines being executed.
type ExecutionInMemory struct {
	Execution

	Events                  []event.EventLogImpl `json:"events"`
	LastProcessedEventIndex int
}

func GetExecution(executionID string) (*ExecutionInMemory, error) {
	exCached, found := cache.GetCache().Get(executionID)
	if !found {
		slog.Debug("Execution not found in cache", "execution_id", executionID)
		return nil, perr.NotFoundWithMessage("Execution " + executionID + " not found")
	}

	ex, ok := exCached.(*ExecutionInMemory)
	if !ok {
		slog.Error("Error casting execution from cache", "execution_id", executionID)
		return nil, perr.InternalWithMessage("Error casting execution " + executionID + " from cache")
	}

	return ex, nil
}

func completeExecution(executionID string) error {
	ex, err := GetExecution(executionID)
	if err != nil && !perr.IsNotFound(err) {
		slog.Error("Error getting execution from cache to complete execution", "execution_id", executionID, "error", err)
		return err
	} else if perr.IsNotFound(err) {
		return nil
	}

	// Leave in cache for 10 minutes
	ok := cache.GetCache().SetWithTTL(executionID, ex, 10*time.Minute)
	if !ok {
		slog.Error("Error setting execution in cache", "execution_id", executionID)
		return perr.InternalWithMessage("Error setting execution in cache")
	}
	return nil
}

func GetPipelineDefnFromExecution(executionID, pipelineExecutionID string) (*ExecutionInMemory, *modconfig.Pipeline, error) {
	ex, err := GetExecution(executionID)
	if err != nil {
		return nil, nil, err
	}
	defn, err := ex.PipelineDefinition(pipelineExecutionID)
	if err != nil {
		return nil, nil, err
	}

	return ex, defn, nil
}

func (ex *ExecutionInMemory) IsPaused() bool {
	paused := false
	for _, pe := range ex.PipelineExecutions {
		if pe.IsPaused() {
			paused = true
		} else {
			paused = false
			break
		}
	}

	return paused
}

func (ex *ExecutionInMemory) EndExecution() error {
	// This seems a convenient place to expire the execution from the cache
	err := completeExecution(ex.ID)
	if err != nil {
		slog.Error("Error completing execution", "error", err)
		return err
	}

	return nil
}

func (ex *ExecutionInMemory) AddEvent(evt event.EventLogImpl) error {
	ex.Events = append(ex.Events, evt)
	err := ex.ProcessEvents()
	return err
}

func (ex *ExecutionInMemory) BuildEvalContext(pipelineDefn *modconfig.Pipeline, pe *PipelineExecution) (*hcl.EvalContext, error) {
	executionVariables, err := pe.GetExecutionVariables()
	if err != nil {
		return nil, err
	}

	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}

	params := map[string]cty.Value{}

	// Why do we add notifier earlier? Because of the param validation before
	notifierMap, err := parse.BuildNotifierMapForEvalContext(fpConfig.Notifiers)
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypeNotifier] = cty.ObjectVal(notifierMap)

	// **temporarily** add add connections to eval context .. we need to remove them later and only add connections
	// that are used by the pipelines. The connections are special because they may need to be resolved before
	// we use them i.e. temp AWS creds.

	connMap := parse.BuildTemporaryConnectionMapForEvalContext(fpConfig.PipelingConnections)
	evalContext.Variables[schema.BlockTypeConnection] = cty.ObjectVal(connMap)

	for _, v := range pipelineDefn.Params {
		if pe.Args[v.Name] != nil {
			paramArg := pe.Args[v.Name]

			if !hclhelpers.IsComplexType(v.Type) && !v.Type.HasDynamicTypes() {
				val, err := gocty.ToCtyValue(pe.Args[v.Name], v.Type)
				if err != nil {
					return nil, err
				}
				params[v.Name] = val
			} else if mapParam, ok := paramArg.(map[string]any); ok && mapParam["resource_type"] == schema.BlockTypeNotifier {

				// Special handling for Notifier type param. Connection type param is different, it has late binding requirement
				// but notifier can be fully resolved here
				// find the notifier in the fpConfig.Notifiers
				notifierName := mapParam["name"].(string)
				notifier, ok := fpConfig.Notifiers[notifierName]
				if !ok {
					return nil, perr.BadRequestWithMessage("notifier not found: " + notifierName)
				}
				notifierCtyValue, err := notifier.CtyValue()
				if err != nil {
					return nil, err
				}

				params[v.Name] = notifierCtyValue

			} else {
				// we'll do our best here
				val, err := hclhelpers.ConvertInterfaceToCtyValue(pe.Args[v.Name])
				if err != nil {
					return nil, err
				}
				params[v.Name] = val
			}

		} else {
			params[v.Name] = v.Default
		}

		// validate pipeline param
		//
		// One of the validation is the "subtype" validation. There are only 2 subtypes supported:
		// 1. connection
		// 2. notifier
		//
		validParam, diags, err := v.ValidateSetting(params[v.Name], evalContext)
		if err != nil {
			slog.Error("Failed to validate pipeline param", "error", err)
			return nil, err
		}

		if !validParam {
			if len(diags) > 0 {
				return nil, error_helpers.BetterHclDiagsToError(v.Name, diags)
			}
			return nil, perr.BadRequestWithMessage("invalid value for param " + v.Name)
		}
	}

	if log.GetLogLevel().Level() == slog.LevelDebug {
		for k, p := range params {
			goVal, err := hclhelpers.CtyToGo(p)
			if err != nil {
				slog.Debug("Error converting cty value to Go value", "error", err)
			} else {
				slog.Debug("Parsed param value", "name", k, "value", goVal)
			}
		}
	}

	paramsCtyVal := cty.ObjectVal(params)
	evalContext.Variables[schema.BlockTypeParam] = paramsCtyVal

	pipelineMap, err := ex.buildPipelineMapForEvalContext()
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypePipeline] = cty.ObjectVal(pipelineMap)

	integrationMap, err := buildIntegrationMapForEvalContext()
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypeIntegration] = cty.ObjectVal(integrationMap)

	// populate the variables and locals
	// build a variables map _excluding_ late binding vars, and a separate map for late binding vars
	variablesMap, _, lateBindingVarDeps := parse.VariableValueCtyMap(pipelineDefn.GetMod().ResourceMaps.Variables, true)

	// add these to eval context
	evalContext.Variables[constants.LateBindingVarsKey] = cty.ObjectVal(lateBindingVarDeps)
	for _, variable := range pipelineDefn.GetMod().ResourceMaps.Variables {
		variablesMap[variable.ShortName] = variable.Value
	}
	evalContext.Variables[schema.AttributeVar] = cty.ObjectVal(variablesMap)

	localsMap := make(map[string]cty.Value)
	for _, local := range pipelineDefn.GetMod().ResourceMaps.Locals {
		localsMap[local.ShortName] = local.Value
	}
	evalContext.Variables[schema.AttributeLocal] = cty.ObjectVal(localsMap)

	// get the nested mod resource (just the pipelines for now)
	if pipelineDefn.GetMod().HasDependentMods() {
		for _, dependentMod := range pipelineDefn.GetMod().ResourceMaps.Mods {
			if dependentMod.Name() == pipelineDefn.GetMod().Name() {
				continue
			}

			nestedModResources, err := buildNestedModResourcesForEvalContext(dependentMod)
			if err != nil {
				return nil, err
			}

			evalContext.Variables[dependentMod.ShortName] = nestedModResources
		}
	}

	return evalContext, nil
}

// This function mutates evalContext
func (ex *ExecutionInMemory) AddCredentialsToEvalContext(evalContext *hcl.EvalContext, stepDefn modconfig.PipelineStep) (*hcl.EvalContext, error) {

	// We should NOT add all credentials in EvalContext, this is why it's done a bit complicated. Credentials need to be resolved, and some (AWS) resolution
	// can be expensive, i.e. getting session token. So we try to "guess" which credentials are required. It's not perfect, especially when the credentials
	// need to be resolved at runtime. This is because we don't know what the value would be until we run the step.
	//
	// If you look at the following function (buildCredentialMapForEvalContext) it tries to guess the credentials that it may need to resolve
	if stepDefn != nil && len(stepDefn.GetCredentialDependsOn()) > 0 {
		params := map[string]cty.Value{}

		if evalContext.Variables[schema.BlockTypeParam] != cty.NilVal {
			params = evalContext.Variables[schema.BlockTypeParam].AsValueMap()
		}

		credentialMap, err := ex.buildCredentialMapForEvalContext(stepDefn.GetCredentialDependsOn(), params)
		if err != nil {
			return nil, err
		}

		// Override what we have
		evalContext.Variables[schema.BlockTypeCredential] = cty.ObjectVal(credentialMap)
	}

	return evalContext, nil
}

func (ex *ExecutionInMemory) AddCredentialsToEvalContextFromPipeline(evalContext *hcl.EvalContext, pipelineDefn *modconfig.Pipeline) (*hcl.EvalContext, error) {
	stepDefns := pipelineDefn.Steps

	allCredentialsDependsOn := []string{}
	for _, stepDefn := range stepDefns {
		allCredentialsDependsOn = append(allCredentialsDependsOn, stepDefn.GetCredentialDependsOn()...)
	}

	pipelineOutputs := pipelineDefn.OutputConfig
	for _, output := range pipelineOutputs {
		allCredentialsDependsOn = append(allCredentialsDependsOn, output.CredentialDependsOn...)
	}

	params := map[string]cty.Value{}
	if evalContext.Variables[schema.BlockTypeParam] != cty.NilVal {
		params = evalContext.Variables[schema.BlockTypeParam].AsValueMap()
	}

	credentialMap, err := ex.buildCredentialMapForEvalContext(allCredentialsDependsOn, params)
	if err != nil {
		return nil, err
	}

	// Override what we have
	evalContext.Variables[schema.BlockTypeCredential] = cty.ObjectVal(credentialMap)

	return evalContext, nil
}

func (ex *ExecutionInMemory) AddConnectionsToEvalContextFromPipeline(evalContext *hcl.EvalContext, pipelineDefn *modconfig.Pipeline) (*hcl.EvalContext, error) {
	stepDefns := pipelineDefn.Steps

	allConnectionsDependsOn := []string{}
	for _, stepDefn := range stepDefns {
		allConnectionsDependsOn = append(allConnectionsDependsOn, stepDefn.GetConnectionDependsOn()...)
	}

	pipelineOutputs := pipelineDefn.OutputConfig
	for _, output := range pipelineOutputs {
		allConnectionsDependsOn = append(allConnectionsDependsOn, output.ConnectionDependsOn...)
	}

	params := map[string]cty.Value{}
	if evalContext.Variables[schema.BlockTypeParam] != cty.NilVal {
		params = evalContext.Variables[schema.BlockTypeParam].AsValueMap()
	}

	vars := map[string]cty.Value{}
	if evalContext.Variables["var"] != cty.NilVal {
		vars = evalContext.Variables["var"].AsValueMap()
	}

	connectionMap, newParamsMap, varMap, err := BuildConnectionMapForEvalContext(allConnectionsDependsOn, params, vars, pipelineDefn.Params)
	if err != nil {
		return nil, err
	}

	// Override what we have
	evalContext.Variables[schema.BlockTypeConnection] = cty.ObjectVal(connectionMap)
	evalContext.Variables[schema.BlockTypeParam] = cty.ObjectVal(newParamsMap)
	evalContext.Variables[schema.AttributeVar] = cty.ObjectVal(varMap)

	return evalContext, nil
}

func (ex *ExecutionInMemory) buildCredentialMapForEvalContext(credentialsInContext []string, params map[string]cty.Value) (map[string]cty.Value, error) {
	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}

	allCredentials := fpConfig.Credentials
	relevantCredentials := map[string]credential.Credential{}
	dynamicCredsType := map[string]bool{}

	for _, credentialName := range credentialsInContext {
		if allCredentials[credentialName] != nil {
			relevantCredentials[credentialName] = allCredentials[credentialName]
		}

		if strings.Contains(credentialName, "<dynamic>") {
			parts := strings.Split(credentialName, ".")
			if len(parts) > 0 {
				dynamicCredsType[parts[0]] = true
			}
		}
	}

	if len(dynamicCredsType) > 0 {
		for _, v := range params {
			// Determine if the credential "may" be needed based on the param value.
			if v.Type() == cty.String && !v.IsNull() {
				potentialCredName := v.AsString()
				for _, c := range allCredentials {
					if c.GetHclResourceImpl().ShortName == potentialCredName && dynamicCredsType[c.GetCredentialType()] {
						relevantCredentials[c.Name()] = c
						break
					}
				}
			}
		}
	}

	credentialMap, err := evalCredentialMapForEvalContext(context.TODO(), relevantCredentials)
	if err != nil {
		return nil, err
	}

	return credentialMap, nil
}

func (ex *ExecutionInMemory) buildPipelineMapForEvalContext() (map[string]cty.Value, error) {
	allPipelines, err := db.ListAllPipelines()
	if err != nil {
		return nil, err
	}

	return buildPipelineMap(allPipelines)
}

func buildPipelineMap(allPipelines []*modconfig.Pipeline) (map[string]cty.Value, error) {

	pipelineMap := map[string]cty.Value{}
	for _, p := range allPipelines {

		// TODO: this doesn't work with mods
		parts := strings.Split(p.Name(), ".")
		if len(parts) != 3 {
			return nil, perr.BadRequestWithMessage("invalid pipeline name: " + p.Name())
		}

		pCty, err := p.CtyValue()
		if err != nil {
			return nil, err
		}
		pipelineMap[parts[2]] = pCty
	}

	return pipelineMap, nil
}

func buildNestedModResourcesForEvalContext(nestedMod *modconfig.Mod) (cty.Value, error) {

	allPipelines := []*modconfig.Pipeline{}
	for _, r := range nestedMod.ResourceMaps.Pipelines {
		if r.ModName != nestedMod.ShortName {
			continue
		}

		allPipelines = append(allPipelines, r)
	}

	pipelineMap, err := buildPipelineMap(allPipelines)
	if err != nil {
		return cty.NilVal, err
	}

	nestedModResources := cty.ObjectVal(
		map[string]cty.Value{
			"pipeline": cty.ObjectVal(pipelineMap),
		},
	)

	return nestedModResources, nil
}

func buildIntegrationMapForEvalContext() (map[string]cty.Value, error) {
	integrationMap := map[string]cty.Value{}
	slackIntegrationMap := map[string]cty.Value{}
	emailIntegrationMap := map[string]cty.Value{}
	teamsIntegrationMap := map[string]cty.Value{}

	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}

	for _, p := range fpConfig.Integrations {

		parts := strings.Split(p.Name(), ".")

		if len(parts) != 2 {
			return nil, perr.BadRequestWithMessage("invalid integration name: " + p.Name())
		}

		pCty, err := p.CtyValue()
		if err != nil {
			return nil, err
		}

		integrationType := parts[0]

		switch integrationType {
		case schema.IntegrationTypeSlack:
			slackIntegrationMap[parts[1]] = pCty
		case schema.IntegrationTypeEmail:
			emailIntegrationMap[parts[1]] = pCty
		case schema.IntegrationTypeMsTeams:
			teamsIntegrationMap[parts[1]] = pCty
		case schema.IntegrationTypeHttp:
			// do nothing

		default:
			return nil, perr.BadRequestWithMessage("invalid integration type: " + integrationType)
		}
	}

	if len(slackIntegrationMap) > 0 {
		integrationMap[schema.IntegrationTypeSlack] = cty.ObjectVal(slackIntegrationMap)
	}

	if len(emailIntegrationMap) > 0 {
		integrationMap[schema.IntegrationTypeEmail] = cty.ObjectVal(emailIntegrationMap)
	}

	if len(teamsIntegrationMap) > 0 {
		integrationMap[schema.IntegrationTypeMsTeams] = cty.ObjectVal(teamsIntegrationMap)
	}

	return integrationMap, nil

}

// StepDefinition returns the step definition for the given step execution ID.
func (ex *ExecutionInMemory) StepDefinition(pipelineExecutionID, stepExecutionID string) (modconfig.PipelineStep, error) {

	pe := ex.PipelineExecutions[pipelineExecutionID]
	se, ok := pe.StepExecutions[stepExecutionID]

	if !ok {
		return nil, perr.BadRequestWithMessage("step execution not found: " + stepExecutionID)
	}
	pd, err := ex.PipelineDefinition(se.PipelineExecutionID)
	if err != nil {
		return nil, err
	}
	sd := pd.GetStep(se.Name)
	if helpers.IsNil(sd) {
		return nil, perr.InternalWithMessage("mod definition may have changed since execution, step '" + se.Name + "' not found")
	}
	return sd, nil
}

func (ex *ExecutionInMemory) PipelineDefinition(pipelineExecutionID string) (*modconfig.Pipeline, error) {

	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, perr.BadRequestWithMessage("pipeline execution " + pipelineExecutionID + " not found")
	}

	pipeline, err := db.GetPipeline(pe.Name)

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (ex *ExecutionInMemory) PipelineData(pipelineExecutionID string) (map[string]interface{}, error) {

	// Get the outputs from prior steps in the pipeline
	data, err := ex.PipelineStepOutputs(pipelineExecutionID)
	if err != nil {
		return nil, err
	}

	// Add arguments data for this pipeline execution
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]

	if !ok {
		return nil, perr.BadRequestWithMessage("pipeline execution not found: " + pipelineExecutionID)
	}

	// Arguments data takes precedence over a step output with the same name
	data[schema.AttributeTypeArgs] = pe.Args

	// TODO - Add variables data for this pipeline execution
	return data, nil
}

// PipelineStepOutputs returns a single map of all outputs from all steps in
// the given pipeline execution. The map is keyed by the step name. If a step
// has a ForTemplate then the result is an array of outputs.
func (ex *ExecutionInMemory) PipelineStepOutputs(pipelineExecutionID string) (map[string]interface{}, error) {

	pe := ex.PipelineExecutions[pipelineExecutionID]

	outputs := map[string]interface{}{}
	for _, se := range pe.StepExecutions {
		if se.PipelineExecutionID != pipelineExecutionID {
			continue
		}
		if _, ok := outputs[se.Name]; !ok {
			outputs[se.Name] = []interface{}{}
		}
		outputs[se.Name] = append(outputs[se.Name].([]interface{}), se.Output)
	}
	return outputs, nil
}

// ParentStepExecution returns the parent step execution for the given pipeline
// execution ID.
func (ex *ExecutionInMemory) ParentStepExecution(pipelineExecutionID string) (*StepExecution, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, fmt.Errorf("pipeline execution %s not found", pipelineExecutionID)
	}
	if pe.ParentStepExecutionID == "" {
		return nil, nil
	}

	parentPe, ok := ex.PipelineExecutions[pe.ParentExecutionID]
	if !ok {
		return nil, fmt.Errorf("parent pipeline execution not found: %s", pe.ParentStepExecutionID)
	}

	se, ok := parentPe.StepExecutions[pe.ParentStepExecutionID]
	if !ok {
		return nil, fmt.Errorf("parent step execution not found: %s", pe.ParentStepExecutionID)
	}
	return se, nil
}

func (ex *ExecutionInMemory) PipelineStepExecutions(pipelineExecutionID, stepName string) []StepExecution {
	pe := ex.PipelineExecutions[pipelineExecutionID]

	results := []StepExecution{}

	for _, se := range pe.StepExecutions {
		results = append(results, *se)
	}

	return results
}

func (ex *ExecutionInMemory) ProcessEvents() error {
	for i := ex.LastProcessedEventIndex; i < len(ex.Events); i++ {
		event := ex.Events[i]
		err := ex.AppendEventLogEntry(event)
		if err != nil {
			slog.Error("Fail to append event entry to execution", "execution", ex.ID, "error", err, "event", event)
			return err
		}
	}
	ex.LastProcessedEventIndex = len(ex.Events)

	return nil
}

func (ex *ExecutionInMemory) AppendSerialisedEventLogEntry(logEntry event.EventLogImpl) error {

	// logEntry.Detail is a map[string]interface{} because we just read it from the database
	//
	// we can do something smarter later check if it's interface{} or fully formed struct (not sure if the use case will appear later)
	jsonData, err := json.Marshal(logEntry.GetDetail())
	if err != nil {
		slog.Error("Fail to marshal event detail", "execution", ex.ID, "error", err)
		return perr.InternalWithMessage("Fail to marshal event detail")
	}

	switch logEntry.GetEventType() {
	case ExecutionQueuedEvent.HandlerName(): // "handler.execution_queued"
		var et event.ExecutionQueued
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.execution_queued event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_queued event")
		}

		return ex.appendEvent(&et)

	case ExecutionStartedEvent.HandlerName(): // "handler.execution_started"
		var et event.ExecutionStarted
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.execution_started event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_started event")
		}

		return ex.appendEvent(&et)

	case ExecutionFinishedEvent.HandlerName(): // "handler.execution_finished"
		var et event.ExecutionFinished
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.execution_finished event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_finished event")
		}

		return ex.appendEvent(&et)

	case ExecutionFailedEvent.HandlerName(): // "handler.execution_failed"
		var et event.ExecutionFailed
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.execution_failed event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_failed event")
		}

		return ex.appendEvent(&et)

	case PipelineQueueCommand.HandlerName(): // "command.pipeline_queue"
		var et event.PipelineQueue
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall command.pipeline_queue event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_queue event")
		}

		return ex.appendEvent(&et)

	case PipelineQueuedEvent.HandlerName(): // "handler.pipeline_queued"
		var et event.PipelineQueued
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_queued event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_queued event")
		}

		return ex.appendEvent(&et)

	case PipelineStartedEvent.HandlerName(): // "handler.pipeline_started"
		var et event.PipelineStarted

		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_started event")
		}

		return ex.appendEvent(&et)

	case PipelineResumedEvent.HandlerName(): // "handler.pipeline_resumed"
		var et event.PipelineStarted
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_resumed event")
		}

		return ex.appendEvent(&et)

	case PipelinePlannedEvent.HandlerName(): // "handler.pipeline_planned"
		var et event.PipelinePlanned
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_planned event")
		}

		return ex.appendEvent(&et)

	case StepQueueCommand.HandlerName(): //  "command.step_queue"
		var et event.StepQueue
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall command.step_queue event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall command.step_queue event")
		}

		db.MapStepExecutionID(logEntry.ProcessID, et.PipelineExecutionID, et.StepExecutionID)

		return ex.appendEvent(&et)

	case StepStartCommand.HandlerName(): // "command.step_start"
		var et event.StepStart
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall command.step_start event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall command.step_start event")
		}

		return ex.appendEvent(&et)

	case StepPipelineStartedEvent.HandlerName(): //  "handler.step_pipeline_started"
		var et event.StepPipelineStarted
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.step_pipeline_started event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_pipeline_started event")
		}

		return ex.appendEvent(&et)

	case StepFinishedEvent.HandlerName(): //  "handler.step_finished"
		var et event.StepFinished
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.step_finished event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_finished event")
		}

		return ex.appendEvent(&et)

	case StepForEachPlannedEvent.HandlerName(): // "handler.step_for_each_planned"
		var et event.StepForEachPlanned
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.step_for_each_planned event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_for_each_planned event")
		}

		return ex.appendEvent(&et)

	case PipelineCanceledEvent.HandlerName(): // "handler.pipeline_canceled"
		var et event.PipelineCanceled
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_canceled event")
		}

		return ex.appendEvent(&et)

	case PipelinePausedEvent.HandlerName(): //  "handler.pipeline_paused"
		var et event.PipelinePaused
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_paused event")
		}

		return ex.appendEvent(&et)

	case PipelineFinishCommand.HandlerName(): // "command.pipeline_finish"
		var et event.PipelineFinished
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_finish event")
		}

		return ex.appendEvent(&et)

	case PipelineFinishedEvent.HandlerName(): // "handler.pipeline_finished"
		var et event.PipelineFinished
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_finished event")
		}

		return ex.appendEvent(&et)

	case PipelineFailedEvent.HandlerName(): // "handler.pipeline_failed"
		var et event.PipelineFailed
		err := json.Unmarshal(jsonData, &et)
		if err != nil {
			slog.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID, "error", err)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_failed event")
		}

		return ex.appendEvent(&et)

	default:
		// TODO: should we ignore unknown types or error out?
	}

	return nil
}

func (ex *ExecutionInMemory) AppendEventLogEntry(logEntry event.EventLogImpl) error {

	switch logEntry.GetEventType() {

	case ExecutionQueuedEvent.HandlerName(): // "handler.execution_queued"
		et, ok := logEntry.GetDetail().(*event.ExecutionQueued)
		if !ok {
			slog.Error("Fail to unmarshall handler.execution_queued event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_queued event")
		}

		return ex.appendEvent(et)

	case ExecutionStartedEvent.HandlerName(): // "handler.execution_started"
		et, ok := logEntry.GetDetail().(*event.ExecutionStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.execution_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_started event")
		}

		return ex.appendEvent(et)

	case ExecutionFinishedEvent.HandlerName(): // "handler.execution_finished"
		et, ok := logEntry.GetDetail().(*event.ExecutionFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.execution_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_finished event")
		}

		return ex.appendEvent(et)

	case ExecutionFailedEvent.HandlerName(): // "handler.execution_failed"
		et, ok := logEntry.GetDetail().(*event.ExecutionFailed)
		if !ok {
			slog.Error("Fail to unmarshall handler.execution_failed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.execution_failed event")
		}

		return ex.appendEvent(et)

	case PipelineQueueCommand.HandlerName(): // "command.pipeline_queue"
		et, ok := logEntry.GetDetail().(*event.PipelineQueue)
		if !ok {
			slog.Error("Fail to unmarshall command.pipeline_queue event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_queue event")
		}

		return ex.appendEvent(et)

	case PipelineQueuedEvent.HandlerName(): // "handler.pipeline_queued"
		et, ok := logEntry.GetDetail().(*event.PipelineQueued)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_queued event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_queued event")
		}

		return ex.appendEvent(et)

	case PipelineStartedEvent.HandlerName(): // "handler.pipeline_started"
		et, ok := logEntry.GetDetail().(*event.PipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_started event")
		}

		return ex.appendEvent(et)

	case PipelineResumedEvent.HandlerName(): // "handler.pipeline_resumed"
		et, ok := logEntry.GetDetail().(*event.PipelineResumed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_resumed event")
		}

		return ex.appendEvent(et)

	case PipelinePlannedEvent.HandlerName(): // "handler.pipeline_planned"
		et, ok := logEntry.GetDetail().(*event.PipelinePlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_planned event")
		}

		return ex.appendEvent(et)

	case StepQueueCommand.HandlerName(): //  "command.step_queue"
		et, ok := logEntry.GetDetail().(*event.StepQueue)
		if !ok {
			slog.Error("Fail to unmarshall command.step_queue event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_queue event")
		}

		return ex.appendEvent(et)

	case StepStartCommand.HandlerName(): // "command.step_start"
		et, ok := logEntry.GetDetail().(*event.StepStart)
		if !ok {
			slog.Error("Fail to unmarshall command.step_start event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_start event")
		}

		return ex.appendEvent(et)

	case StepPipelineStartedEvent.HandlerName(): //  "handler.step_pipeline_started"
		et, ok := logEntry.GetDetail().(*event.StepPipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_pipeline_started event")
		}

		return ex.appendEvent(et)

	// this is the generic step finish event that is fired by the command.step_start command
	case StepFinishedEvent.HandlerName(): //  "handler.step_finished"
		et, ok := logEntry.GetDetail().(*event.StepFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_finished event")
		}

		return ex.appendEvent(et)

	case StepForEachPlannedEvent.HandlerName(): // "handler.step_for_each_planned"
		et, ok := logEntry.GetDetail().(*event.StepForEachPlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_for_each_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_for_each_planned event")
		}

		return ex.appendEvent(et)

	case PipelineCanceledEvent.HandlerName(): // "handler.pipeline_canceled"
		et, ok := logEntry.GetDetail().(*event.PipelineCanceled)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_canceled event")
		}

		return ex.appendEvent(et)

	case PipelinePausedEvent.HandlerName(): //  "handler.pipeline_paused"
		et, ok := logEntry.GetDetail().(*event.PipelinePaused)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_paused event")
		}

		return ex.appendEvent(et)

	case PipelineFinishCommand.HandlerName(): // "command.pipeline_finish"
		et, ok := logEntry.GetDetail().(*event.PipelineFinish)
		if !ok {
			slog.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_finish event")
		}

		return ex.appendEvent(et)

	case PipelineFinishedEvent.HandlerName(): // "handler.pipeline_finished"
		et, ok := logEntry.GetDetail().(*event.PipelineFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_finished event")
		}

		return ex.appendEvent(et)

	case PipelineFailedEvent.HandlerName(): // "handler.pipeline_failed"
		et, ok := logEntry.GetDetail().(*event.PipelineFailed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_failed event")
		}

		return ex.appendEvent(et)

	default:
		// TODO: should we ignore unknown types or error out?
	}

	return nil
}

func SaveEventToSQLite(db *sql.DB, executionID string, event event.EventLogImpl) error {
	retentionInSecond := viper.GetInt(constants.ArgProcessRetention)
	if retentionInSecond == 0 {
		return nil
	}

	payloadData, err := json.Marshal(event.GetDetail())
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)
		return err
	}

	sanitizedPayloadData := sanitize.Instance.SanitizeString(string(payloadData))

	statement := `insert into event (id, struct_version, process_id, created_at, message, level, detail) values (?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(statement, event.GetID(), event.GetStructVersion(), executionID, event.GetCreatedAt(), event.GetEventType(), event.GetLevel(), sanitizedPayloadData)
	if err != nil {
		return err
	}
	return nil
}

func LoadExecutionFromProcessDB(e *event.Event) (*ExecutionInMemory, error) {

	if e.ExecutionID == "" {
		return nil, perr.BadRequestWithMessage("event execution ID is empty")
	}

	ex := &ExecutionInMemory{
		Execution: Execution{
			ID:                 e.ExecutionID,
			PipelineExecutions: map[string]*PipelineExecution{},
		},
	}

	var localLock *sync.Mutex
	if ex.Lock == nil {
		localLock = event.GetEventStoreMutex(e.ExecutionID)
		localLock.Lock()
		defer func() {
			if localLock != nil {
				localLock.Unlock()
			}
		}()
	}

	db, err := store.OpenFlowpipeDB()
	if err != nil {
		return nil, err
	}

	// Prepare query to select all events
	query := `select id, struct_version, process_id, message, level, created_at, detail from event where process_id = ? order by created_at asc`
	rows, err := db.Query(query, e.ExecutionID)
	if err != nil {
		slog.Error("error querying event table", "error", err)
		return nil, perr.InternalWithMessage("error querying event table")
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {

		var id string
		var structVersion string
		var processId string
		var message string
		var level string
		var createdAt time.Time
		var detailString string

		err := rows.Scan(&id, &structVersion, &processId, &message, &level, &createdAt, &detailString)
		if err != nil {
			slog.Error("error scanning event table", "error", err)
			return nil, perr.InternalWithMessage("error scanning event table")
		}

		ele := event.EventLogImpl{
			ID:            id,
			StructVersion: structVersion,
			ProcessID:     processId,
			Message:       message,
			Level:         level,
			CreatedAt:     createdAt,
		}

		// marshall the payload
		var detail interface{}
		err = json.Unmarshal([]byte(detailString), &detail)
		if err != nil {
			slog.Error("error unmarshalling event payload", "error", err)
			return nil, perr.InternalWithMessage("error unmarshalling event payload")
		}

		ele.SetDetail(detail)

		err = ex.AppendSerialisedEventLogEntry(ele)
		if err != nil {
			slog.Error("Fail to append event entry to execution", "execution", ex.ID, "error", err, "string", detailString)
			return nil, err
		}
	}

	if rows.Err() != nil {
		slog.Error("error iterating event table", "error", rows.Err())
		return nil, perr.InternalWithMessage("error iterating event table")
	}

	return ex, nil
}
