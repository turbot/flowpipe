package execution

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/types"
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

	// Steps triggered by pipelines in the execution. We maintain a flat list
	// of all steps triggered by all pipelines for easy lookup and querying.
	StepExecutions map[string]*StepExecution `json:"step_executions"`
}

// ExecutionOption is a function that modifies an Execution instance.
type ExecutionOption func(*Execution) error

func NewExecution(ctx context.Context, opts ...ExecutionOption) (*Execution, error) {

	ex := &Execution{
		// ID is empty by default, so it will be populated from the given event
		Context:            ctx,
		PipelineExecutions: map[string]*PipelineExecution{},
		StepExecutions:     map[string]*StepExecution{},
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
func (ex *Execution) StepDefinition(stepExecutionID string) (*types.PipelineStep, error) {
	se, ok := ex.StepExecutions[stepExecutionID]
	if !ok {
		return nil, fmt.Errorf("step execution %s not found", stepExecutionID)
	}
	pd, err := ex.PipelineDefinition(se.PipelineExecutionID)
	if err != nil {
		return nil, err
	}
	sd := pd.Steps[se.Name]
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
		return nil, fmt.Errorf("pipeline execution %s not found", pipelineExecutionID)
	}
	// Arguments data takes precedence over a step output with the same name
	data["args"] = pe.Args

	// TODO - Add variables data for this pipeline execution

	return data, nil
}

// PipelineStepOutputs returns a single map of all outputs from all steps in
// the given pipeline execution. The map is keyed by the step name. If a step
// has a ForTemplate then the result is an array of outputs.
func (ex *Execution) PipelineStepOutputs(pipelineExecutionID string) (map[string]interface{}, error) {
	outputs := map[string]interface{}{}
	for stepExecutionID, se := range ex.StepExecutions {
		if se.PipelineExecutionID != pipelineExecutionID {
			continue
		}
		sd, err := ex.StepDefinition(stepExecutionID)
		if err != nil {
			return nil, err
		}
		if sd.For == "" {
			outputs[se.Name] = se.Output
		} else {
			if _, ok := outputs[se.Name]; !ok {
				outputs[se.Name] = []interface{}{}
			}
			outputs[se.Name] = append(outputs[se.Name].([]interface{}), se.Output)
		}
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
	se, ok := ex.StepExecutions[pe.ParentStepExecutionID]
	if !ok {
		return nil, fmt.Errorf("step execution %s not found", pe.ParentStepExecutionID)
	}
	return se, nil
}

// PipelineStepExecutions returns a list of step executions for the given
// pipeline execution ID and step name.
func (ex *Execution) PipelineStepExecutions(pipelineExecutionID, stepName string) []StepExecution {
	var stepExecutions []StepExecution
	for _, se := range ex.StepExecutions {
		if se.PipelineExecutionID == pipelineExecutionID && se.Name == stepName {
			stepExecutions = append(stepExecutions, *se)
		}
	}
	return stepExecutions
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
		return fperr.BadRequestWithMessage("event execution ID is empty")
	}

	if ex.ID == "" {
		ex.ID = e.ExecutionID
	}

	if ex.ID != e.ExecutionID {
		return fperr.BadRequestWithMessage("event execution ID (" + e.ExecutionID + ") does not match execution ID (" + ex.ID + ")")
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
			return err
		}

		logger.Trace("<1> event type #3", "eventType", ele.EventType)
		switch ele.EventType {
		case "handler.pipeline_queued":
			var et event.PipelineQueued
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			ex.PipelineExecutions[et.PipelineExecutionID] = &PipelineExecution{
				ID:                    et.PipelineExecutionID,
				Name:                  et.Name,
				Args:                  et.Args,
				Status:                "queued",
				StepStatus:            map[string]*StepStatus{},
				ParentStepExecutionID: et.ParentStepExecutionID,
			}

		case "handler.pipeline_started":
			var et event.PipelineStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "started"

		case "handler.pipeline_resumed":
			var et event.PipelineStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			// TODO: is this right?
			pe.Status = "started"

		case "handler.pipeline_planned":
			var et event.PipelinePlanned
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pd, err := ex.PipelineDefinition(et.PipelineExecutionID)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			for _, step := range pd.Steps {
				pe.InitializeStep(step.Name)
			}

		case "command.pipeline_step_start":
			var et event.PipelineStepStart
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			ex.StepExecutions[et.StepExecutionID] = &StepExecution{
				PipelineExecutionID: et.PipelineExecutionID,
				ID:                  et.StepExecutionID,
				Name:                et.StepName,
				Status:              "starting",
			}
			// Set the overall step status
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			stepDefn, err := ex.StepDefinition(et.StepExecutionID)
			if err != nil {
				return err
			}
			ex.StepExecutions[et.StepExecutionID].Input = et.StepInput
			ex.StepExecutions[et.StepExecutionID].ForEach = et.ForEach
			pe.StepStatus[stepDefn.Name].Queue(et.StepExecutionID)

		case "handler.pipeline_step_started":
			var et event.PipelineStepStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			// Step the specific step execution status
			ex.StepExecutions[et.StepExecutionID].Status = "started"
			stepDefn, err := ex.StepDefinition(et.StepExecutionID)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.StartStep(stepDefn.Name, et.StepExecutionID)

		case "handler.pipeline_step_finished":
			var et event.PipelineStepFinished
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			stepDefn, err := ex.StepDefinition(et.StepExecutionID)
			if err != nil {
				return err
			}
			// Step the specific step execution status
			ex.StepExecutions[et.StepExecutionID].Status = "finished"
			ex.StepExecutions[et.StepExecutionID].Output = et.Output

			// TODO: ignore error setting -> we need to be able to ignore setting
			// TODO: is a step failure an immediate end of the pipeline?
			// TODO: can a pipeline continue if a step fails? Is that the ignore setting?
			if et.Error != nil {
				ex.StepExecutions[et.StepExecutionID].Error = et.Error
				logger.Trace("Setting pipeline step finish error", "stepExecutionID", et.StepExecutionID, "error", et.Error)
				ex.StepExecutions[et.StepExecutionID].Status = "failed"
				pe.FailStep(stepDefn.Name, et.StepExecutionID)
			} else {
				pe.FinishStep(stepDefn.Name, et.StepExecutionID)
			}

		case "handler.pipeline_canceled":
			var et event.PipelineCanceled
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "canceled"

		case "handler.pipeline_paused":
			var et event.PipelinePaused
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "paused"

		case "handler.pipeline_finished":
			var et event.PipelineFinished
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "finished"
			pe.Output = et.Output

		case "handler.pipeline_failed":
			var et event.PipelineFailed
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "failed"

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
