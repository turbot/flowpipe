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
	"github.com/turbot/flowpipe/pipeparser/funcs"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
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
			val, err := gocty.ToCtyValue(pe.Args[k], v.Type)
			if err != nil {
				return nil, err
			}
			params[k] = val
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

		pipelineMap[parts[2]] = p.AsCtyValue()
	}

	evalContext.Variables[schema.BlockTypePipeline] = cty.ObjectVal(pipelineMap)

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
func (ex *Execution) StepDefinition(pipelineExecutionID, stepExecutionID string) (modconfig.IPipelineStep, error) {
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

	// Find the step execution order first
	orders := pe.StepExecutionOrder[stepName]
	if len(orders) == 0 {
		// TODO: Error?
		return nil
	}

	results := make([]StepExecution, len(orders))

	for i, stepExecutionID := range orders {
		se := pe.StepExecutions[stepExecutionID]
		results[i] = *se
	}
	return results
}

// LogFilePath returns the path to the log file for the execution.
func (ex *Execution) LogFilePath() (string, error) {
	filename := fmt.Sprintf("%s.jsonl", ex.ID)
	p := filepath.Join(viper.GetString("log.dir"), filename)
	return filepath.Abs(p)
}

// This function loads the event log file (the .jsonl file) continously and update the
// ex.PipelineExecutions and ex.StepExecutions
func (ex *Execution) LoadProcess(e *event.Event) error {

	logger := fplog.Logger(ex.Context)

	logger.Trace("<1> execution.LoadProcess #1", "executionID", ex.ID, "event executionID", e.ExecutionID)

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
	logger.Trace("<1> Loading file #2", "execution", ex.ID, "logPath", logPath)

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
				StepStatus:            map[string]*StepStatus{},
				ParentStepExecutionID: et.ParentStepExecutionID,
				ParentExecutionID:     et.ParentExecutionID,
				Errors:                []modconfig.StepError{},
				AllNativeStepOutputs:  ExecutionStepOutputs{},
				AllConfigStepOutputs:  ExecutionStepOutputs{},
				StepExecutions:        map[string]*StepExecution{},
				StepExecutionOrder:    map[string][]string{},
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
			pe.StepExecutionOrder[et.StepName] = append(pe.StepExecutionOrder[et.StepName], et.StepExecutionID)

			stepDefn, err := ex.StepDefinition(et.PipelineExecutionID, et.StepExecutionID)
			if err != nil {
				logger.Error("Failed to get step definition - 1", "execution", ex.ID, "stepExecutionID", et.StepExecutionID, "error", err)
				return err
			}
			pe.StepExecutions[et.StepExecutionID].Input = et.StepInput
			pe.StepExecutions[et.StepExecutionID].StepForEach = et.StepForEach
			pe.StepExecutions[et.StepExecutionID].NextStepAction = et.NextStepAction
			pe.StepStatus[stepDefn.GetFullyQualifiedName()].Queue(et.StepExecutionID)

		case "command.pipeline_step_start":
			var et event.PipelineStepStart
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				logger.Error("Fail to unmarshall command.pipeline_step_start event", "execution", ex.ID, "error", err)
				return err
			}

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

			pe.StartStep(stepDefn.GetFullyQualifiedName(), et.StepExecutionID)

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
				logger.Error("Failed to get step definition - 3", "stepExecutionID", et.StepExecutionID, "error", err)
				return err
			}

			shouldBeIndexed := false
			if stepDefn.GetForEach() != nil {
				shouldBeIndexed = true
			}

			// Step the specific step execution status
			if pe.StepExecutions[et.StepExecutionID] == nil {
				return perr.BadRequestWithMessage("Unable to find step execution " + et.StepExecutionID + " in pipeline execution " + pe.ID)
			}

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

			if pe.AllNativeStepOutputs[stepDefn.GetType()] == nil {
				pe.AllNativeStepOutputs[stepDefn.GetType()] = map[string]interface{}{}
			}

			if pe.AllConfigStepOutputs[stepDefn.GetType()] == nil {
				pe.AllConfigStepOutputs[stepDefn.GetType()] = map[string]interface{}{}
			}

			if !shouldBeIndexed {
				// non for_each step. The step will be accessed such as:
				// text = step.echo.text_1.text
				pe.AllNativeStepOutputs[stepDefn.GetType()][stepDefn.GetName()] = et.Output

				pe.AllConfigStepOutputs[stepDefn.GetType()][stepDefn.GetName()] = et.StepOutput
			} else {
				// for indexed step, you want to be able to access the step as
				// text = step.echo.text_1[1].text

				if pe.AllNativeStepOutputs[stepDefn.GetType()][stepDefn.GetName()] == nil {
					pe.AllNativeStepOutputs[stepDefn.GetType()][stepDefn.GetName()] = make([]*modconfig.Output, et.StepForEach.TotalCount)
				}

				pe.AllNativeStepOutputs[stepDefn.GetType()][stepDefn.GetName()].([]*modconfig.Output)[et.StepForEach.Index] = et.Output

				if pe.AllConfigStepOutputs[stepDefn.GetType()][stepDefn.GetName()] == nil {
					pe.AllConfigStepOutputs[stepDefn.GetType()][stepDefn.GetName()] = make([]map[string]interface{}, et.StepForEach.TotalCount)
				}

				pe.AllConfigStepOutputs[stepDefn.GetType()][stepDefn.GetName()].([]map[string]interface{})[et.StepForEach.Index] = et.StepOutput
			}

			// TODO: Error handling
			// TODO: ignore error setting -> we need to be able to ignore setting
			// TODO: is a step failure an immediate end of the pipeline?
			// TODO: can a pipeline continue if a step fails? Is that the ignore setting?
			if et.Output.HasErrors() {

				// TODO: ignore retries for now (StepFinalFailure)
				if !stepDefn.GetErrorConfig().Ignore {
					// pe.StepExecutions[et.StepExecutionID].Error = et.Error
					// logger.Trace("Setting pipeline step finish error", "stepExecutionID", et.StepExecutionID, "error", et.Error)
					// pe.StepExecutions[et.StepExecutionID].Status = "failed"
					pe.FailStep(stepDefn.GetFullyQualifiedName(), et.StepExecutionID)
					pe.Fail(stepDefn.GetFullyQualifiedName(), et.Output.Errors...)
				} else {
					// Should we add the step errors to PipelineExecution.Errors if the error is ignored?
					pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepExecutionID)
				}

				// TODO: this below comment is not true anymore, keep this here until we refactor how we handle the failure & retries
				// IMPORTANT: we must call this to check if this step is the final failure
				// this function also sets the internal error tracker of the pe. Not sure if that's right place
				// to do it

				// stepFinalFailure := pe.IsStepFinalFailure(stepDefn, ex)
				// if stepFinalFailure {
				// 	logger.Trace("Step final failure", "step", stepDefn)
				// }
			} else {
				pe.FinishStep(stepDefn.GetFullyQualifiedName(), et.StepExecutionID)
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
