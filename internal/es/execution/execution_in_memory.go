package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/pipe-fittings/constants"
	pfconstants "github.com/turbot/pipe-fittings/constants"
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
		slog.Error("Error getting execution from cache", "execution_id", executionID)
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
		Functions: funcs.ContextFunctions(viper.GetString(pfconstants.ArgModLocation)),
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

	connMap, err := parse.BuildTemporaryConnectionMapForEvalContext(context.TODO(), fpConfig.PipelingConnections)
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypeConnection] = cty.ObjectVal(connMap)

	for _, v := range pipelineDefn.Params {
		if pe.Args[v.Name] != nil {
			if !v.Type.HasDynamicTypes() && !v.Type.IsCapsuleType() {
				val, err := gocty.ToCtyValue(pe.Args[v.Name], v.Type)
				if err != nil {
					return nil, err
				}
				params[v.Name] = val
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
	variablesMap := make(map[string]cty.Value)
	for _, variable := range pipelineDefn.GetMod().ResourceMaps.Variables {
		variablesMap[variable.ShortName] = variable.Value
	}
	evalContext.Variables[schema.AttributeVar] = cty.ObjectVal(variablesMap)

	localsMap := make(map[string]cty.Value)
	for _, local := range pipelineDefn.GetMod().ResourceMaps.Locals {
		localsMap[local.ShortName] = local.Value
	}
	evalContext.Variables[schema.AttributeLocal] = cty.ObjectVal(localsMap)

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

	connectionMap, newParamsMap, err := ex.buildConnectionMapForEvalContext(allConnectionsDependsOn, params, pipelineDefn)
	if err != nil {
		return nil, err
	}

	// Override what we have
	evalContext.Variables[schema.BlockTypeConnection] = cty.ObjectVal(connectionMap)
	evalContext.Variables[schema.BlockTypeParam] = cty.ObjectVal(newParamsMap)

	return evalContext, nil
}

func (ex *ExecutionInMemory) buildConnectionMapForEvalContext(connectionsInContext []string, params map[string]cty.Value, pipelineDefn *modconfig.Pipeline) (map[string]cty.Value, map[string]cty.Value, error) {
	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, nil, err
	}

	allConnections := fpConfig.PipelingConnections
	relevantConnections := map[string]modconfig.PipelingConnection{}
	dynamicConnType := map[string]bool{}

	for _, connectionName := range connectionsInContext {
		if allConnections[connectionName] != nil {
			relevantConnections[connectionName] = allConnections[connectionName]
		}

		if strings.Contains(connectionName, "<dynamic>") {
			parts := strings.Split(connectionName, ".")
			if len(parts) > 0 {
				dynamicConnType[parts[0]] = true
			}
		}
	}

	if len(dynamicConnType) > 0 {
		for _, v := range params {
			// Determine if the credential "may" be needed based on the param value.
			if v.Type() == cty.String && !v.IsNull() {
				potentialConnName := v.AsString()
				for _, c := range allConnections {
					if c.GetHclResourceImpl().ShortName == potentialConnName && dynamicConnType[c.GetConnectionType()] {
						relevantConnections[c.Name()] = c
						break
					}
				}
			}
		}
	}

	for _, p := range pipelineDefn.Params {
		if p.IsCustomType() {
			for k, v := range params {
				if k != p.Name {
					continue
				}

				if v.Type().IsObjectType() || v.Type().IsMapType() {
					valueMap := v.AsValueMap()
					paramToUpdate := extractConnection(valueMap, allConnections, relevantConnections)
					ctyVal, err := paramToUpdate.CtyValue()
					if err != nil {
						return nil, nil, err
					}

					params[p.Name] = ctyVal
					break
				} else if hclhelpers.IsCollectionOrTuple(v.Type()) {
					for _, val := range v.AsValueSlice() {
						valueMap := val.AsValueMap()
						paramToUpdate := extractConnection(valueMap, allConnections, relevantConnections)

						ctyVal, err := paramToUpdate.CtyValue()
						if err != nil {
							return nil, nil, err
						}

						params[p.Name] = ctyVal
					}
				}
			}
		}
	}

	connectionMap, err := evaluateConnectionMapForEvalContext(context.TODO(), relevantConnections)
	if err != nil {
		return nil, nil, err
	}

	return connectionMap, params, nil
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

func (ex *ExecutionInMemory) AppendEventLogEntry(logEntry event.EventLogImpl) error {

	switch logEntry.GetEventType() {

	case PipelineQueuedEvent.HandlerName(): // "handler.pipeline_queued"
		et, ok := logEntry.GetDetail().(*event.PipelineQueued)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_queued event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_queued event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineStartedEvent.HandlerName(): // "handler.pipeline_started"
		et, ok := logEntry.GetDetail().(*event.PipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_started event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineResumedEvent.HandlerName(): // "handler.pipeline_resumed"
		et, ok := logEntry.GetDetail().(*event.PipelineResumed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_resumed event")
		}

		return ex.Execution.appendEvent(et)

	case PipelinePlannedEvent.HandlerName(): // "handler.pipeline_planned"
		et, ok := logEntry.GetDetail().(*event.PipelinePlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_planned event")
		}

		return ex.Execution.appendEvent(et)

	case StepQueueCommand.HandlerName(): //  "command.step_queue"
		et, ok := logEntry.GetDetail().(*event.StepQueue)
		if !ok {
			slog.Error("Fail to unmarshall command.step_queue event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_queue event")
		}

		return ex.Execution.appendEvent(et)

	case StepStartCommand.HandlerName(): // "command.step_start"
		et, ok := logEntry.GetDetail().(*event.StepStart)
		if !ok {
			slog.Error("Fail to unmarshall command.step_start event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_start event")
		}

		return ex.Execution.appendEvent(et)

	case StepPipelineStartedEvent.HandlerName(): //  "handler.step_pipeline_started"
		et, ok := logEntry.GetDetail().(*event.StepPipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_pipeline_started event")
		}

		return ex.Execution.appendEvent(et)

	// this is the generic step finish event that is fired by the command.step_start command
	case StepFinishedEvent.HandlerName(): //  "handler.step_finished"
		et, ok := logEntry.GetDetail().(*event.StepFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_finished event")
		}

		return ex.Execution.appendEvent(et)

	case StepForEachPlannedEvent.HandlerName(): // "handler.step_for_each_planned"
		et, ok := logEntry.GetDetail().(*event.StepForEachPlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_for_each_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_for_each_planned event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineCanceledEvent.HandlerName(): // "handler.pipeline_canceled"
		et, ok := logEntry.GetDetail().(*event.PipelineCanceled)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_canceled event")
		}

		return ex.Execution.appendEvent(et)

	case PipelinePausedEvent.HandlerName(): //  "handler.pipeline_paused"
		et, ok := logEntry.GetDetail().(*event.PipelinePaused)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_paused event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineFinishCommand.HandlerName(): // "command.pipeline_finish"
		et, ok := logEntry.GetDetail().(*event.PipelineFinish)
		if !ok {
			slog.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_finish event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineFinishedEvent.HandlerName(): // "handler.pipeline_finished"
		et, ok := logEntry.GetDetail().(*event.PipelineFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_finished event")
		}

		return ex.Execution.appendEvent(et)

	case PipelineFailedEvent.HandlerName(): // "handler.pipeline_failed"
		et, ok := logEntry.GetDetail().(*event.PipelineFailed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_failed event")
		}

		return ex.Execution.appendEvent(et)

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
