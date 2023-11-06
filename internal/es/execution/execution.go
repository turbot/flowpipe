package execution

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
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

	evalContext.Variables[schema.BlockTypePipeline] = cty.ObjectVal(pipelineMap)

	integrationMap := map[string]cty.Value{}
	slackIntegrationMap := map[string]cty.Value{}
	emailIntegrationMap := map[string]cty.Value{}

	for _, p := range pipelineDefn.GetMod().ResourceMaps.Integrations {

		parts := strings.Split(p.Name(), ".")
		if len(parts) != 4 {
			return nil, perr.BadRequestWithMessage("invalid integration name: " + p.Name())
		}

		integrationType := parts[2]
		switch integrationType {
		case string(schema.IntegrationTypeSlack):
			slackIntegration := p.(*modconfig.SlackIntegration)
			pCty, err := slackIntegration.CtyValue()
			if err != nil {
				return nil, err
			}
			slackIntegrationMap[parts[3]] = pCty

		case string(schema.IntegrationTypeEmail):
			emailIntegration := p.(*modconfig.EmailIntegration)
			pCty, err := emailIntegration.CtyValue()
			if err != nil {
				return nil, err
			}
			emailIntegrationMap[parts[3]] = pCty
		}
	}

	if len(slackIntegrationMap) > 0 {
		integrationMap[schema.IntegrationTypeSlack] = cty.ObjectVal(slackIntegrationMap)
	}

	if len(emailIntegrationMap) > 0 {
		integrationMap[schema.IntegrationTypeEmail] = cty.ObjectVal(emailIntegrationMap)
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
	p := filepath.Join(viper.GetString(constants.ArgLogDir), filename)
	return filepath.Abs(p)
}

// This function loads the event log file (the .jsonl file) continously and update the
// ex.PipelineExecutions and ex.StepExecutions
func (ex *Execution) LoadProcess(e *event.Event) error {

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
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize*20), bufio.MaxScanTokenSize*20)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var ele types.EventLogEntry
		err := json.Unmarshal(ba, &ele)
		if err != nil {
			logger.Error("Fail to unmarshall event log entry", "execution", ex.ID, "error", err, "string", string(ba))
			return err
		}

		switch ele.EventType {
		case "handler.pipeline_queued":
			var et event.PipelineQueued
			err := json.Unmarshal(ele.Payload, &et)
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

		case "handler.pipeline_started":
			var et event.PipelineStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_started event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "started"

		case "handler.pipeline_resumed":
			var et event.PipelineStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_resumed event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			// TODO: is this right?
			pe.Status = "started"

		case "handler.pipeline_planned":
			var et event.PipelinePlanned
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_planned event", "execution", ex.ID, "error", err)
				return err
			}
			pd, err := ex.PipelineDefinition(et.PipelineExecutionID)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			for _, step := range pd.Steps {
				pe.InitializeStep(step.GetFullyQualifiedName())
			}

		// TODO: I'm not sure if this is the right move. Initially I was using this to introduce the concept of a "queue"
		// TODO: for the step (just like we're queueing the pipeline). But I'm not sure if it's really required, we could just
		// TODO: delay the start. We need to evolve this as we go.
		case "command.pipeline_step_queue":
			var et event.PipelineStepStart
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall command.pipeline_step_queue event", "execution", ex.ID, "error", err)
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

		case "command.pipeline_step_start":
			var et event.PipelineStepStart
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall command.pipeline_step_start event", "execution", ex.ID, "error", err)
				return err
			}

		// handler.pipeline_step_started is the event when the pipeline is starting a child pipeline, i.e. "pipeline step", this isn't
		// a generic step start event
		case "handler.pipeline_step_started":
			var et event.PipelineStepStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_step_started event", "execution", ex.ID, "error", err)
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

		// this is the generic step finish event that is fired by the command.pipeline_step_start command
		case "handler.pipeline_step_finished":
			var et event.PipelineStepFinished
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_step_finished event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			stepDefn, err := ex.StepDefinition(pe.ID, et.StepExecutionID)
			if err != nil {
				logger.Error("Failed to get step definition", "stepExecutionID", et.StepExecutionID, "error", err)
				return err
			}

			loopContinue := false
			if et.StepLoop != nil {
				if !et.StepLoop.LoopCompleted {
					loopContinue = true
				}
			}

			// Step the specific step execution status
			if pe.StepExecutions[et.StepExecutionID] == nil {
				return perr.BadRequestWithMessage("Unable to find step execution " + et.StepExecutionID + " in pipeline execution " + pe.ID)
			}

			// pe.StepExecutions[et.StepExecutionID].StepForEach should be set at the beginning of the step execution, not here
			// StepLoop on the other hand, can only be determined at the end of the step, so this is the right place to do it
			pe.StepExecutions[et.StepExecutionID].StepLoop = et.StepLoop

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
					pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopContinue)
				}
			} else {
				pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepForEach.Key, et.StepExecutionID, loopContinue)
			}

		case "handler.pipeline_canceled":
			var et event.PipelineCanceled
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_canceled event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "canceled"

		case "handler.pipeline_paused":
			var et event.PipelinePaused
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_paused event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "paused"

		case "command.pipeline_finish":
			var et event.PipelineFinished
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall command.pipeline_finish event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "finishing"

		case "handler.pipeline_finished":
			var et event.PipelineFinished
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_finished event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "finished"
			pe.PipelineOutput = et.PipelineOutput

		case "handler.pipeline_failed":
			var et event.PipelineFailed
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall handler.pipeline_failed event", "execution", ex.ID, "error", err)
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "failed"
			pe.PipelineOutput = et.PipelineOutput
			if et.Error != nil {
				if pe.Errors == nil {
					pe.Errors = []modconfig.StepError{}
				}
				pe.Errors = append(pe.Errors, *et.Error)
			}

		default:
			// Ignore unknown types while loading
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
