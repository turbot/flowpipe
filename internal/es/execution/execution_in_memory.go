package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/sanitize"
	pfconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Execution represents the current state of an execution. A single execution
// is tied to a trigger (webhook, cronjob, etc) and may result in multiple
// pipelines being executed.
type ExecutionInMemory struct {
	// Unique identifier for this execution.
	ID string `json:"id"`

	// Pipelines triggered by the execution. Even if the pipelines are nested,
	// we maintain a flat list of all pipelines for easy lookup and querying.
	PipelineExecutions map[string]*PipelineExecution `json:"pipeline_executions"`

	Events                  []event.EventLogEntry `json:"events"`
	LastProcessedEventIndex int
	Lock                    *sync.Mutex `json:"-"`
}

func GetExecution(executionID string) (*ExecutionInMemory, error) {
	exCached, found := cache.GetCache().Get(executionID)
	if !found {
		slog.Error("Error getting execution from cache", "execution_id", executionID)
		return nil, perr.NotFoundWithMessage("Execution " + executionID + " not found")
	}

	ex, ok := exCached.(*ExecutionInMemory)
	if !ok {
		slog.Error("Error casting execution from cache", "execution_id", executionID)
		return nil, perr.InternalWithMessage("Error casting execution " + executionID + " from cache")
	}

	return ex, nil
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

func (ex *ExecutionInMemory) SaveToFile() error {
	eventStoreFilePath := filepaths.EventStoreFilePath(ex.ID)

	stat, _ := os.Stat(eventStoreFilePath)
	if stat != nil {
		// Keeping this simple, we don't want to overwrite the file. This may change in the future when we can resume execution.
		// Right now if Flowpipe stops/crashes there's no way to resume the execution because it's no longer in memory and not
		// persisted
		return perr.BadRequestWithMessage("execution file already exists. execution can only be serialised once at termination")
	}

	// Append the JSON data to a file
	file, err := os.OpenFile(eventStoreFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return perr.InternalWithMessage("Error opening file " + err.Error())
	}
	defer file.Close()

	for _, event := range ex.Events {
		// Marshal the struct to JSON
		eventData, err := json.Marshal(event) // No indent, single line
		if err != nil {
			slog.Error("Error marshalling JSON", "error", err)
			return err
		}

		sanitizedEventData := sanitize.Instance.SanitizeString(string(eventData))

		_, err = file.Write([]byte(sanitizedEventData))
		if err != nil {
			return perr.InternalWithMessage("Error writing to file " + err.Error())
		}

		_, err = file.WriteString("\n")
		if err != nil {
			return perr.InternalWithMessage("Error writing to file " + err.Error())
		}
	}
	return nil
}

func (ex *ExecutionInMemory) AddEvent(evt event.EventLogEntry) error {
	ex.Events = append(ex.Events, evt)
	err := ex.ProcessEvents()
	return err
}

func (ex *ExecutionInMemory) BuildEvalContext(pipelineDefn *modconfig.Pipeline, pe *PipelineExecution) (*hcl.EvalContext, error) {
	executionVariables, err := pe.GetExecutionVariables()
	if err != nil {
		return nil, err
	}

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(pfconstants.ArgModLocation)),
	}

	params := map[string]cty.Value{}

	for k, v := range pipelineDefn.Params {
		if pe.Args[k] != nil {
			if !v.Type.HasDynamicTypes() {
				val, err := gocty.ToCtyValue(pe.Args[k], v.Type)
				if err != nil {
					return nil, err
				}
				params[k] = val
			} else {
				// we'll do our best here
				val, err := hclhelpers.ConvertInterfaceToCtyValue(pe.Args[k])
				if err != nil {
					return nil, err
				}
				params[k] = val
			}

		} else {
			params[k] = v.Default
		}
	}

	paramsCtyVal := cty.ObjectVal(params)
	evalContext.Variables[schema.BlockTypeParam] = paramsCtyVal

	pipelineMap, err := ex.buildPipelineMapForEvalContext()
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypePipeline] = cty.ObjectVal(pipelineMap)

	integrationMap, err := ex.buildIntegrationMapForEvalContext(pipelineDefn)
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

func (ex *ExecutionInMemory) buildCredentialMapForEvalContext(credentialsInContext []string, params map[string]cty.Value) (map[string]cty.Value, error) {
	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}

	allCredentials := fpConfig.Credentials
	relevantCredentials := map[string]modconfig.Credential{}
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

	credentialMap, err := buildCredentialMapForEvalContext(context.TODO(), relevantCredentials)
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

func (ex *ExecutionInMemory) buildIntegrationMapForEvalContext(pipelineDefn *modconfig.Pipeline) (map[string]cty.Value, error) {
	integrationMap := map[string]cty.Value{}
	slackIntegrationMap := map[string]cty.Value{}
	emailIntegrationMap := map[string]cty.Value{}

	for _, p := range pipelineDefn.GetMod().ResourceMaps.Integrations {

		parts := strings.Split(p.Name(), ".")
		if len(parts) != 4 {
			return nil, perr.BadRequestWithMessage("invalid integration name: " + p.Name())
		}

		pCty, err := p.CtyValue()
		if err != nil {
			return nil, err
		}

		integrationType := parts[2]

		switch integrationType {
		case string(schema.IntegrationTypeSlack):
			slackIntegrationMap[parts[3]] = pCty

		case string(schema.IntegrationTypeEmail):
			emailIntegrationMap[parts[3]] = pCty

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
	// Do not attempt to lock, the calling function must orchestrate the locking

	for i := ex.LastProcessedEventIndex; i < len(ex.Events); i++ {
		event := ex.Events[i]
		err := ex.AppendEventLogEntry(event)
		if err != nil {
			slog.Error("Fail to append event log entry to execution", "execution", ex.ID, "error", err, "event", event)
			return err
		}
	}
	ex.LastProcessedEventIndex = len(ex.Events)

	return nil
}

// LoadFromFile loads an execution from a JSON file.
func (ex *ExecutionInMemory) LoadJSON(fileName string) error {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, &ex)
	if err != nil {
		return err
	}
	return nil
}

func (ex *ExecutionInMemory) AppendEventLogEntry(logEntry event.EventLogEntry) error {

	switch logEntry.EventType {

	case PipelineQueuedEvent.HandlerName(): // "handler.pipeline_queued"
		et, ok := logEntry.Payload.(*event.PipelineQueued)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_queued event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_queued event")
		}

		ex.PipelineExecutions[et.PipelineExecutionID] = &PipelineExecution{
			ID:                    et.PipelineExecutionID,
			Name:                  et.Name,
			Args:                  et.Args,
			Status:                "queued",
			StepStatus:            map[string]map[string]*StepStatus{},
			ParentStepExecutionID: et.ParentStepExecutionID,
			ParentExecutionID:     et.ParentExecutionID,
			Errors:                []modconfig.StepError{},
			StepExecutions:        map[string]*StepExecution{},
		}

	case PipelineStartedEvent.HandlerName(): // "handler.pipeline_started"
		et, ok := logEntry.Payload.(*event.PipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_started event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "started"
		pe.StartTime = et.Event.CreatedAt

	case PipelineResumedEvent.HandlerName(): // "handler.pipeline_resumed"
		et, ok := logEntry.Payload.(*event.PipelineResumed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_resumed event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		// TODO: is this right?
		pe.Status = "started"

	case PipelinePlanCommand.HandlerName(): // "command.pipeline_plan"
		_, ok := logEntry.Payload.(*event.PipelinePlan)
		if !ok {
			slog.Error("Fail to unmarshall command.pipeline_plan event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_plan event")
		}

	case PipelinePlannedEvent.HandlerName(): // "handler.pipeline_planned"
		et, ok := logEntry.Payload.(*event.PipelinePlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_planned event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]

		for _, nextStep := range et.NextSteps {
			pe.InitializeStep(nextStep.StepName)
		}

	// TODO: I'm not sure if this is the right move. Initially I was using this to introduce the concept of a "queue"
	// TODO: for the step (just like we're queueing the pipeline). But I'm not sure if it's really required, we could just
	// TODO: delay the start. We need to evolve this as we go.
	case StepQueueCommand.HandlerName(): //  "command.step_queue"
		et, ok := logEntry.Payload.(*event.StepQueue)
		if !ok {
			slog.Error("Fail to unmarshall command.step_queue event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_queue event")
		}

		// Set the overall step status
		pe := ex.PipelineExecutions[et.PipelineExecutionID]

		pe.StepExecutions[et.StepExecutionID] = &StepExecution{
			PipelineExecutionID: et.PipelineExecutionID,
			ID:                  et.StepExecutionID,
			Name:                et.StepName,
			Status:              "starting",
		}

		stepDefn, err := ex.StepDefinition(et.PipelineExecutionID, et.StepExecutionID)
		if err != nil {
			slog.Error("Failed to get step definition - 1", "execution", ex.ID, "stepExecutionID", et.StepExecutionID, "error", err)
			return err
		}
		pe.StepExecutions[et.StepExecutionID].Input = et.StepInput
		pe.StepExecutions[et.StepExecutionID].StepForEach = et.StepForEach
		pe.StepExecutions[et.StepExecutionID].NextStepAction = et.NextStepAction

		if pe.StepStatus[stepDefn.GetFullyQualifiedName()] == nil {
			pe.StepStatus[stepDefn.GetFullyQualifiedName()] = map[string]*StepStatus{}
		}

		if pe.StepStatus[stepDefn.GetFullyQualifiedName()][et.StepForEach.Key] == nil {
			pe.StepStatus[stepDefn.GetFullyQualifiedName()][et.StepForEach.Key] = &StepStatus{
				Queued:   map[string]bool{},
				Started:  map[string]bool{},
				Finished: map[string]bool{},
				Failed:   map[string]bool{},
			}
		}

		pe.StepStatus[stepDefn.GetFullyQualifiedName()][et.StepForEach.Key].Queue(et.StepExecutionID)

	case StepQueuedEvent.HandlerName(): // "handler.step_queued"
		_, ok := logEntry.Payload.(*event.StepQueued)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_queued event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_queued event")
		}

	case StepStartCommand.HandlerName(): // "command.step_start"
		et, ok := logEntry.Payload.(*event.StepStart)
		if !ok {
			slog.Error("Fail to unmarshall command.step_start event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.step_start event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.StepExecutions[et.StepExecutionID].StartTime = et.Event.CreatedAt
		pe.StepExecutions[et.StepExecutionID].StepLoop = et.StepLoop
		pe.StepExecutions[et.StepExecutionID].StepRetry = et.StepRetry

	// handler.step_pipeline_started is the event when the pipeline is starting a child pipeline, i.e. "pipeline step", this isn't
	// a generic step start event
	case StepPipelineStartedEvent.HandlerName(): //  "handler.step_pipeline_started"
		et, ok := logEntry.Payload.(*event.StepPipelineStarted)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_pipeline_started event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_pipeline_started event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]

		// Step the specific step execution status
		pe.StepExecutions[et.StepExecutionID].Status = "started"
		stepDefn, err := ex.StepDefinition(pe.ID, et.StepExecutionID)
		if err != nil {
			slog.Error("Failed to get step definition - 2", "stepExecutionID", et.StepExecutionID, "error", err)
			return err
		}

		pe.StartStep(stepDefn.GetFullyQualifiedName(), et.Key, et.StepExecutionID)
		pe.StepExecutions[et.StepExecutionID].StartTime = et.Event.CreatedAt

	// this is the generic step finish event that is fired by the command.step_start command
	case StepFinishedEvent.HandlerName(): //  "handler.step_finished"
		et, ok := logEntry.Payload.(*event.StepFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_finished event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		stepDefn, err := ex.StepDefinition(pe.ID, et.StepExecutionID)
		if err != nil {
			slog.Error("Failed to get step definition", "stepExecutionID", et.StepExecutionID, "error", err)
			return err
		}

		loopHold := false
		if et.StepLoop != nil && !et.StepLoop.LoopCompleted {
			loopHold = true
		}

		errorHold := false
		if et.StepRetry != nil && !et.StepRetry.RetryCompleted {
			errorHold = true
		}

		// Step the specific step execution status
		if pe.StepExecutions[et.StepExecutionID] == nil {
			return perr.BadRequestWithMessage("Unable to find step execution " + et.StepExecutionID + " in pipeline execution " + pe.ID)
		}

		// pe.StepExecutions[et.StepExecutionID].StepForEach should be set at the beginning of the step execution, not here
		// StepLoop on the other hand, can only be determined at the end of the step. The "LoopCompleted" and "RetryCompleted"
		// are calculated at the end of the step, so we need to overwrite whatever the StepLoop and StepRetry that we have in the beginning
		// of the step execution
		pe.StepExecutions[et.StepExecutionID].StepLoop = et.StepLoop
		pe.StepExecutions[et.StepExecutionID].StepRetry = et.StepRetry

		if et.Output == nil {
			// return fperr.BadRequestWithMessage("Step execution has a nil output " + et.StepExecutionID + " in pipeline execution " + pe.ID)
			slog.Warn("Step execution has a nil output", "stepExecutionID", et.StepExecutionID, "pipelineExecutionID", pe.ID)
		} else {
			pe.StepExecutions[et.StepExecutionID].Status = et.Output.Status
			pe.StepExecutions[et.StepExecutionID].Output = et.Output
		}

		if len(et.StepOutput) > 0 {
			pe.StepExecutions[et.StepExecutionID].StepOutput = et.StepOutput
		}

		pe.StepExecutions[et.StepExecutionID].EndTime = et.Event.CreatedAt

		// TODO: Fix creating duplicate data as we dereference before appending (moved EndTime above this so it is passed into StepStatus)
		// append the Step Execution to the StepStatus (yes it's duplicate data, we may be able to refactor this later)
		pe.StepStatus[stepDefn.GetFullyQualifiedName()][et.StepForEach.Key].StepExecutions = append(pe.StepStatus[stepDefn.GetFullyQualifiedName()][et.StepForEach.Key].StepExecutions,
			*pe.StepExecutions[et.StepExecutionID])

		if et.Output.HasErrors() {
			if et.Output.FailureMode == constants.FailureModeIgnored {
				// Should we add the step errors to PipelineExecution.Errors if the error is ignored?
				pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopHold, errorHold)
			} else {
				pe.FailStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID)

				if !errorHold {
					// if there's a retry config, don't add that failure to the pipeline failure until the final retry attempt
					//
					// retry completed is represented in the errorHold variable
					pe.Fail(stepDefn.GetFullyQualifiedName(), et.Output.Errors...)
				}
			}
		} else {
			pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopHold, errorHold)
		}

	case StepForEachPlannedEvent.HandlerName(): // "handler.step_for_each_planned"
		et, ok := logEntry.Payload.(*event.StepForEachPlanned)
		if !ok {
			slog.Error("Fail to unmarshall handler.step_for_each_planned event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.step_for_each_planned event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		stepStatusMap := pe.StepStatus[et.StepName]

		if len(et.NextSteps) == 0 {
			// this means the for_each step has complete (or failed), mark it as such

			// TODO: I don't think this is the end state
			if len(stepStatusMap) == 0 {
				stepStatusMap["0"] = &StepStatus{
					OverralState: "empty_for_each",
				}
			} else {
				for _, stepStatus := range stepStatusMap {
					stepStatus.OverralState = "complete_or_fail"
				}
			}
		} else {
			for _, v := range et.NextSteps {
				if stepStatusMap[v.StepForEach.Key] == nil {
					stepStatusMap[v.StepForEach.Key] = &StepStatus{
						Initializing: true,
						Queued:       map[string]bool{},
						Started:      map[string]bool{},
						Finished:     map[string]bool{},
						Failed:       map[string]bool{},
					}
				}
			}
		}
		pe.StepStatus[et.StepName] = stepStatusMap

		// if there's NextSteps .. then we assume that the step is still running

	case PipelineCanceledEvent.HandlerName(): // "handler.pipeline_canceled"
		et, ok := logEntry.Payload.(*event.PipelineCanceled)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_canceled event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "canceled"
		pe.EndTime = et.Event.CreatedAt

	case PipelinePausedEvent.HandlerName(): //  "handler.pipeline_paused"
		et, ok := logEntry.Payload.(*event.PipelinePaused)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_paused event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "paused"

	case PipelineFinishCommand.HandlerName(): // "command.pipeline_finish"
		et, ok := logEntry.Payload.(*event.PipelineFinish)
		if !ok {
			slog.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall command.pipeline_finish event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "finishing"

	case PipelineFinishedEvent.HandlerName(): // "handler.pipeline_finished"
		et, ok := logEntry.Payload.(*event.PipelineFinished)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_finished event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "finished"
		pe.EndTime = et.Event.CreatedAt
		pe.PipelineOutput = et.PipelineOutput

	case PipelineFailedEvent.HandlerName(): // "handler.pipeline_failed"

		et, ok := logEntry.Payload.(*event.PipelineFailed)
		if !ok {
			slog.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID)
			return perr.InternalWithMessage("Fail to unmarshall handler.pipeline_failed event")
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = constants.StateFailed
		pe.EndTime = et.Event.CreatedAt
		pe.PipelineOutput = et.PipelineOutput

		if pe.PipelineOutput == nil {
			pe.PipelineOutput = map[string]interface{}{}
		}
		if pe.PipelineOutput["errors"] != nil && len(et.Errors) > 0 {
			for _, e := range et.Errors {

				found := false
				for _, pipelineErr := range pe.PipelineOutput["errors"].([]modconfig.StepError) {
					if e.Error.ID == pipelineErr.Error.ID {
						found = true
						break
					}
				}
				if !found {
					pe.PipelineOutput["errors"] = append(pe.PipelineOutput["errors"].([]modconfig.StepError), et.Errors...)
				}
			}

		} else if pe.PipelineOutput["errors"] == nil && len(et.Errors) > 0 {
			pe.PipelineOutput["errors"] = et.Errors
		}

		// TODO: this is a bit messy
		// pe.Errors are "collected" as we call the pe.Fail() function above during the 'handler.step_finished' handling
		// but **some** thing may call pipeline_failed directly, bypassing the "step_finish" operation (TODO: not sure if this is valid)
		// in that case we need to check et.Errors and "merge" them
		for _, err := range et.Errors {
			found := false
			for _, peErr := range pe.Errors {
				if err.Error.Instance == peErr.Error.Instance {
					found = true
					break
				}
			}
			if !found {
				pe.Errors = append(pe.Errors, err)
			}
		}

	default:
		// Ignore unknown types while loading
	}

	return nil
}
