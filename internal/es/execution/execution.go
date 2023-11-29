package execution

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
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
type Execution struct {
	Context context.Context `json:"-"`

	// Unique identifier for this execution.
	ID string `json:"id"`

	// Pipelines triggered by the execution. Even if the pipelines are nested,
	// we maintain a flat list of all pipelines for easy lookup and querying.
	PipelineExecutions map[string]*PipelineExecution `json:"pipeline_executions"`

	Lock *sync.Mutex `json:"-"`
}

func (ex *Execution) BuildEvalContext(pipelineDefn *modconfig.Pipeline, pe *PipelineExecution) (*hcl.EvalContext, error) {
	executionVariables, err := pe.GetExecutionVariables()
	if err != nil {
		return nil, err
	}

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString("work.dir")),
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
func (ex *Execution) AddCredentialsToEvalContext(evalContext *hcl.EvalContext, stepDefn modconfig.PipelineStep) (*hcl.EvalContext, error) {
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

func (ex *Execution) buildCredentialMapForEvalContext(credentialsInContext []string, params map[string]cty.Value) (map[string]cty.Value, error) {

	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}

	allCredentials := fpConfig.Credentials
	relevantCredentials := map[string]modconfig.Credential{}

	dynamicCredsFound := false
	for _, credentialName := range credentialsInContext {
		if allCredentials[credentialName] != nil {
			relevantCredentials[credentialName] = allCredentials[credentialName]
		}

		if !dynamicCredsFound && strings.Contains(credentialName, "<dynamic>") {
			dynamicCredsFound = true
		}
	}

	if dynamicCredsFound {
		for _, v := range params {
			if v.Type() == cty.String {
				potentialCredName := v.AsString()
				for _, c := range allCredentials {
					if c.GetHclResourceImpl().ShortName == potentialCredName {
						relevantCredentials[c.Name()] = c
						break
					}
				}
			}
		}
	}

	credentialMap, err := buildCredentialMapForEvalContext(relevantCredentials)
	if err != nil {
		return nil, err
	}

	return credentialMap, nil
}

func buildCredentialMapForEvalContext(allCredentials map[string]modconfig.Credential) (map[string]cty.Value, error) {
	credentialMap := map[string]cty.Value{}
	awsCredentialMap := map[string]cty.Value{}
	basicCredentialMap := map[string]cty.Value{}
	slackCedentialMap := map[string]cty.Value{}
	gcpCredentialMap := map[string]cty.Value{}

	cache := cache.GetCache()
	for _, c := range allCredentials {
		parts := strings.Split(c.Name(), ".")
		if len(parts) != 2 {
			return nil, perr.BadRequestWithMessage("invalid credential name: " + c.Name())
		}

		var credToUse modconfig.Credential

		cachedCred, found := cache.Get(c.GetHclResourceImpl().FullName)
		if !found {
			// if not found, call the "resolve" function to resolve this credential, for temp cred this will
			// generate the temp creds
			newC, err := c.Resolve(context.TODO())
			if err != nil {
				return nil, err
			}

			if newC.GetTtl() > 0 {
				cache.SetWithTTL(c.GetHclResourceImpl().FullName, newC, time.Duration(newC.GetTtl())*time.Second)
			}
			credToUse = newC
		} else {
			var ok bool
			credToUse, ok = cachedCred.(modconfig.Credential)
			if !ok {
				return nil, perr.BadRequestWithMessage("invalid credential type: " + c.Name())
			}
		}

		pCty, err := credToUse.CtyValue()
		if err != nil {
			return nil, err
		}

		credentialType := parts[0]

		switch credentialType {
		case "aws":
			awsCredentialMap[parts[1]] = pCty

		case "basic":
			basicCredentialMap[parts[1]] = pCty

		case "slack":
			slackCedentialMap[parts[1]] = pCty

		case "gcp":
			gcpCredentialMap[parts[1]] = pCty

		default:
			return nil, perr.BadRequestWithMessage("invalid credential type: " + credentialType)
		}
	}

	if len(awsCredentialMap) > 0 {
		credentialMap["aws"] = cty.ObjectVal(awsCredentialMap)
	}

	if len(basicCredentialMap) > 0 {
		credentialMap["basic"] = cty.ObjectVal(basicCredentialMap)
	}

	if len(slackCedentialMap) > 0 {
		credentialMap["slack"] = cty.ObjectVal(slackCedentialMap)
	}

	if len(gcpCredentialMap) > 0 {
		credentialMap["gcp"] = cty.ObjectVal(gcpCredentialMap)
	}

	return credentialMap, nil
}

func (ex *Execution) buildPipelineMapForEvalContext() (map[string]cty.Value, error) {
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

func (ex *Execution) buildIntegrationMapForEvalContext(pipelineDefn *modconfig.Pipeline) (map[string]cty.Value, error) {
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

// ExecutionStepOutputs is a map for all the step execution. It's stored in this format:
//
// ExecutionStepOutputs = {
//    "echo" = {
//			"echo_1": {},
//          "my_other_echo": {},
//     },

//	  "http" = {
//	     "http_1": {},
//	     "http_2": {},
//	  }
//	}
//
// The first level is grouping the output by the step type
// The next level group the output by the step name
// The value can be a StepOutput OR a slice of StepOutput
type ExecutionStepOutputs map[string]map[string]interface{}

// ExecutionOption is a function that modifies an Execution instance.
type ExecutionOption func(*Execution) error

func NewExecution(ctx context.Context, opts ...ExecutionOption) (*Execution, error) {

	ex := &Execution{
		// ID is empty by default, so it will be populated from the given event
		Context:            ctx,
		PipelineExecutions: map[string]*PipelineExecution{},
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *Execution as the argument
		err := opt(ex)
		if err != nil {
			return ex, err
		}
	}

	// return the modified execution instance
	return ex, nil

}

// There are only 2 use cases for creator of Execution to provide the lock:
// 1) pipeline planner, and
// 2) step for each planner
//
// # Any other use case we should let the execution object aquire its own lock
//
// NOTE: ensure that WithLock is called *before* WithEvent is called
func WithLock(lock *sync.Mutex) ExecutionOption {
	return func(ex *Execution) error {
		ex.Lock = lock
		return nil
	}
}

func WithID(id string) ExecutionOption {
	return func(ex *Execution) error {
		ex.ID = id
		return nil
	}
}

func WithEvent(e *event.Event) ExecutionOption {
	return func(ex *Execution) error {
		return ex.LoadProcess(e)
	}
}

// StepDefinition returns the step definition for the given step execution ID.
func (ex *Execution) StepDefinition(pipelineExecutionID, stepExecutionID string) (modconfig.PipelineStep, error) {
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
	return sd, nil
}

func (ex *Execution) PipelineData(pipelineExecutionID string) (map[string]interface{}, error) {

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
func (ex *Execution) PipelineStepOutputs(pipelineExecutionID string) (map[string]interface{}, error) {
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
func (ex *Execution) ParentStepExecution(pipelineExecutionID string) (*StepExecution, error) {
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

// PipelineStepExecutions returns a list of step executions for the given
// pipeline execution ID and step name.
func (ex *Execution) PipelineStepExecutions(pipelineExecutionID, stepName string) []StepExecution {
	pe := ex.PipelineExecutions[pipelineExecutionID]

	results := []StepExecution{}

	for _, se := range pe.StepExecutions {
		results = append(results, *se)
	}

	return results
}

// LogFilePath returns the path to the log file for the execution.
func (ex *Execution) LogFilePath() (string, error) {
	filename := fmt.Sprintf("%s.jsonl", ex.ID)
	p := filepath.Join(util.EventStoreDir(), filename)
	return filepath.Abs(p)
}

func getStackTrace() string {
	buf := new(bytes.Buffer)
	for i := 20; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc) //nolint:forbidigo // exception
	}
	return buf.String()
}

// This function loads the event log file (the .jsonl file) continously and update the
// ex.PipelineExecutions and ex.StepExecutions
func (ex *Execution) LoadProcess(e *event.Event) error {

	// Attempt to aquire the execution lock if it's not given. If the execution lock is given then we assume the calling
	// function is responsible for locking & unlocking the event load process.
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

	logger := fplog.Logger(ex.Context)

	if e.ExecutionID == "" {
		return perr.BadRequestWithMessage("event execution ID is empty")
	}

	if ex.ID == "" {
		ex.ID = e.ExecutionID
	}

	if ex.ID != e.ExecutionID {
		return perr.BadRequestWithMessage("event execution ID (" + e.ExecutionID + ") does not match execution ID (" + ex.ID + ")")
	}

	// Open the event log
	logPath, err := ex.LogFilePath()
	if err != nil {
		logger.Error("Failed to get log file path", "execution", ex.ID, "error", err)
		return err
	}

	f, err := os.Open(logPath)
	if err != nil {
		logger.Error("Failed to open log file", "execution", ex.ID, "error", err)
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize*40), bufio.MaxScanTokenSize*40)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var ele types.EventLogEntry
		err := json.Unmarshal(ba, &ele)
		if err != nil {
			stackTrace := getStackTrace()
			logger.Error("Fail to unmarshall event log entry", "execution", ex.ID, "error", err, "string", string(ba), "stackTrace", stackTrace)
			return err
		}

		err = ex.AppendEventLogEntry(ele)
		if err != nil {
			logger.Error("Fail to append event log entry to execution", "execution", ex.ID, "error", err, "string", string(ba))
			return err
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// LoadFromFile loads an execution from a JSON file.
func (ex *Execution) LoadJSON(fileName string) error {
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

// Events
var (
	PipelineQueuedEvent   = event.PipelineQueued{}
	PipelineStartedEvent  = event.PipelineStarted{}
	PipelineResumedEvent  = event.PipelineResumed{}
	PipelinePlannedEvent  = event.PipelinePlanned{}
	PipelineCanceledEvent = event.PipelineCanceled{}
	PipelinePausedEvent   = event.PipelinePaused{}
	PipelineFinishedEvent = event.PipelineFinished{}
	PipelineFailedEvent   = event.PipelineFailed{}
	PipelineLoadedEvent   = event.PipelineLoaded{}

	StepQueuedEvent          = event.StepQueued{}
	StepFinishedEvent        = event.StepFinished{} // this is the generic step finish event that is fired by the command.step_start command
	StepForEachPlannedEvent  = event.StepForEachPlanned{}
	StepPipelineStartedEvent = event.StepPipelineStarted{} // this event is fired for a specific step type: pipeline step (step that launches a pipeline)

)

// Commands
var (
	PipelineCancelCommand = event.PipelineCancel{}
	PipelinePlanCommand   = event.PipelinePlan{}
	PipelineFinishCommand = event.PipelineFinish{}
	PipelineFailCommand   = event.PipelineFail{}
	PipelineLoadCommand   = event.PipelineLoad{}
	PipelinePauseCommand  = event.PipelinePause{}
	PipelineQueueCommand  = event.PipelineQueue{}
	PipelineResumeCommand = event.PipelineResume{}
	PipelineStartCommand  = event.PipelineStart{}

	StepQueueCommand = event.StepQueue{}
	StepStartCommand = event.StepStart{}

	StepPipelineFinishCommand = event.StepPipelineFinish{} // this command is fired when a child pipeline has finished. This is to inform the parent pipeline to continue the execution
)

func (ex *Execution) AppendEventLogEntry(logEntry types.EventLogEntry) error {
	logger := fplog.Logger(ex.Context)

	switch logEntry.EventType {

	case PipelineQueuedEvent.HandlerName(): // "handler.pipeline_queued"
		var et event.PipelineQueued
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_queued event", "execution", ex.ID, "error", err)
			return err
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
		var et event.PipelineStarted
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "started"
		pe.StartTime = et.Event.CreatedAt

	case PipelineResumedEvent.HandlerName(): // "handler.pipeline_resumed"
		var et event.PipelineStarted
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		// TODO: is this right?
		pe.Status = "started"

	case PipelinePlanCommand.HandlerName(): // "command.pipeline_plan"
		var et event.PipelinePlan
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall command.pipeline_plan event", "execution", ex.ID, "error", err)
			return err
		}

	case PipelinePlannedEvent.HandlerName(): // "handler.pipeline_planned"
		var et event.PipelinePlanned
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID, "error", err)
			return err
		}

		pe := ex.PipelineExecutions[et.PipelineExecutionID]

		for _, nextStep := range et.NextSteps {
			pe.InitializeStep(nextStep.StepName)
		}

	// TODO: I'm not sure if this is the right move. Initially I was using this to introduce the concept of a "queue"
	// TODO: for the step (just like we're queueing the pipeline). But I'm not sure if it's really required, we could just
	// TODO: delay the start. We need to evolve this as we go.
	case StepQueueCommand.HandlerName(): //  "command.step_queue"
		var et event.StepStart
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall command.step_queue event", "execution", ex.ID, "error", err)
			return err
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
			logger.Error("Failed to get step definition - 1", "execution", ex.ID, "stepExecutionID", et.StepExecutionID, "error", err)
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
		var et event.StepQueued
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.step_queued event", "execution", ex.ID, "error", err)
			return err
		}

	case StepStartCommand.HandlerName(): // "command.step_start"
		var et event.StepStart
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall command.step_start event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.StepExecutions[et.StepExecutionID].StartTime = et.Event.CreatedAt
		pe.StepExecutions[et.StepExecutionID].StepLoop = et.StepLoop
		pe.StepExecutions[et.StepExecutionID].StepRetry = et.StepRetry

	// handler.step_pipeline_started is the event when the pipeline is starting a child pipeline, i.e. "pipeline step", this isn't
	// a generic step start event
	case StepPipelineStartedEvent.HandlerName(): //  "handler.step_pipeline_started"
		var et event.StepPipelineStarted
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.step_pipeline_started event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]

		// Step the specific step execution status
		pe.StepExecutions[et.StepExecutionID].Status = "started"
		stepDefn, err := ex.StepDefinition(pe.ID, et.StepExecutionID)
		if err != nil {
			logger.Error("Failed to get step definition - 2", "stepExecutionID", et.StepExecutionID, "error", err)
			return err
		}

		pe.StartStep(stepDefn.GetFullyQualifiedName(), et.Key, et.StepExecutionID)
		pe.StepExecutions[et.StepExecutionID].StartTime = et.Event.CreatedAt

	// this is the generic step finish event that is fired by the command.step_start command
	case StepFinishedEvent.HandlerName(): //  "handler.step_finished"
		var et event.StepFinished
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.step_finished event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		stepDefn, err := ex.StepDefinition(pe.ID, et.StepExecutionID)
		if err != nil {
			logger.Error("Failed to get step definition", "stepExecutionID", et.StepExecutionID, "error", err)
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
			logger.Warn("Step execution has a nil output", "stepExecutionID", et.StepExecutionID, "pipelineExecutionID", pe.ID)
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

		// TODO: Error handling
		// TODO: ignore error setting -> we need to be able to ignore setting
		// TODO: is a step failure an immediate end of the pipeline?
		// TODO: can a pipeline continue if a step fails? Is that the ignore setting?
		if et.Output.HasErrors() {
			// TODO: ignore retries for now (StepFinalFailure)
			if !stepDefn.GetErrorConfig().Ignore {
				// pe.StepExecutions[et.StepExecutionID].Error = et.Error
				// pe.StepExecutions[et.StepExecutionID].Status = "failed"
				pe.FailStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID)
				pe.Fail(stepDefn.GetFullyQualifiedName(), et.Output.Errors...)
			} else {
				// Should we add the step errors to PipelineExecution.Errors if the error is ignored?
				pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopHold, errorHold)
			}
		} else {
			pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopHold, errorHold)
		}

	case StepForEachPlannedEvent.HandlerName(): // "handler.step_for_each_planned"
		var et event.StepForEachPlanned
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.step_for_each_planned event", "execution", ex.ID, "error", err)
			return err
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
		var et event.PipelineCanceled
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "canceled"
		pe.EndTime = et.Event.CreatedAt

	case PipelinePausedEvent.HandlerName(): //  "handler.pipeline_paused"
		var et event.PipelinePaused
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "paused"

	case PipelineFinishCommand.HandlerName(): // "command.pipeline_finish"
		var et event.PipelineFinished
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "finishing"

	case PipelineFinishedEvent.HandlerName(): // "handler.pipeline_finished"
		var et event.PipelineFinished
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "finished"
		pe.EndTime = et.Event.CreatedAt
		pe.PipelineOutput = et.PipelineOutput

	case PipelineFailedEvent.HandlerName(): // "handler.pipeline_failed"
		var et event.PipelineFailed
		err := json.Unmarshal(logEntry.Payload, &et)
		if err != nil {
			logger.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID, "error", err)
			return err
		}
		pe := ex.PipelineExecutions[et.PipelineExecutionID]
		pe.Status = "failed"
		pe.EndTime = et.Event.CreatedAt
		pe.PipelineOutput = et.PipelineOutput

		if pe.PipelineOutput == nil {
			pe.PipelineOutput = map[string]interface{}{}
		}
		pe.PipelineOutput["errors"] = et.Errors

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
