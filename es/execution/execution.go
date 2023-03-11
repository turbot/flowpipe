package execution

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe-pipelines/config"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
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
func (ex *Execution) StepDefinition(stepExecutionID string) (*pipeline.PipelineStep, error) {
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

func (ex *Execution) LogFilePath() (string, error) {
	cfg := config.Get(ex.Context)
	filename := fmt.Sprintf("%s.jsonl", ex.ID)
	p := filepath.Join(cfg.LogDir, filename)
	return filepath.Abs(p)
}

func (ex *Execution) LoadProcess(e *event.Event) error {

	if e.ExecutionID == "" {
		return fmt.Errorf("event execution ID is empty: %v", e)
	}

	if ex.ID == "" {
		ex.ID = e.ExecutionID
	}

	if ex.ID != e.ExecutionID {
		return fmt.Errorf("event execution ID (%s) does not match execution ID (%s)", e.ExecutionID, ex.ID)
	}

	// Open the event log
	logPath, err := ex.LogFilePath()
	if err != nil {
		return err
	}
	f, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var ele EventLogEntry
		err := json.Unmarshal(ba, &ele)
		if err != nil {
			return err
		}

		switch ele.EventType {

		case "event.PipelineQueued":
			var et event.PipelineQueued
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			ex.PipelineExecutions[et.PipelineExecutionID] = &PipelineExecution{
				ID:                    et.PipelineExecutionID,
				Name:                  et.Name,
				Input:                 et.Input,
				Status:                "queued",
				StepStatus:            map[string]StepStatus{},
				StepExecutions:        []string{},
				ParentStepExecutionID: et.ParentStepExecutionID,
			}

		case "event.PipelinePlanned":
			var et event.PipelinePlanned
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "planned"
			pd, err := ex.PipelineDefinition(et.PipelineExecutionID)
			if err != nil {
				return err
			}
			for _, nextStep := range et.NextSteps {
				sd := pd.Steps[nextStep]
				queueSize := len(sd.For)
				if queueSize == 0 {
					queueSize = 1
				}
				ex.PipelineExecutions[et.PipelineExecutionID].StepStatus[nextStep] = StepStatus{
					Queued: queueSize,
				}
			}

		case "event.PipelineStarted":
			var et event.PipelineStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			pe := ex.PipelineExecutions[et.PipelineExecutionID]
			pe.Status = "started"

		case "event.PipelineStepStart":
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
			if ss, ok := pe.StepStatus[stepDefn.Name]; ok {
				ss.Queued = ss.Queued - 1
				ss.Started = ss.Started + 1
				pe.StepStatus[stepDefn.Name] = ss
			} else {
				return fmt.Errorf("step %s not found in pipeline %s", stepDefn.Name, et.PipelineExecutionID)
			}

		case "event.PipelineStepStarted":
			var et event.PipelineStepStarted
			err := json.Unmarshal(ele.Payload, &et)
			if err != nil {
				return err
			}
			// Step the specific step execution status
			ex.StepExecutions[et.StepExecutionID].Status = "started"

		case "event.PipelineStepFinished":
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
			if ss, ok := pe.StepStatus[stepDefn.Name]; ok {
				ss.Started = ss.Started - 1
				ss.Finished = ss.Finished + 1
				pe.StepStatus[stepDefn.Name] = ss
			} else {
				return fmt.Errorf("step %s not found in pipeline %s", stepDefn.Name, et.PipelineExecutionID)
			}
			// Step the specific step execution status
			ex.StepExecutions[et.StepExecutionID].Status = "finished"
			ex.StepExecutions[et.StepExecutionID].Output = et.Output

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
	json.Unmarshal([]byte(byteValue), &ex)
	return nil
}

/*

func Dump(ctx context.Context) error {
	s := &State{}
	return s.Load(ctx)
}

// This is an attempt at Loading the state of a session, including the status of
// any running pipelines. Next steps:
// * How do we know the flow of states, ex.g. if load goes to failed (instead of loaded)?
// * How do we inject the right commands / events back into the bus to resume?
func (s *State) Load(ctx context.Context) error {

	logFile := fmt.Sprintf("logs/%s.jsonl", utils.Session(ctx))

	// Open the event log
	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	pipelineStatus := map[string]string{}

	scanner := bufio.NewScanner(f)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var entryMetadata EventLogEntry
		err := json.Unmarshal(ba, &entryMetadata)
		if err != nil {
			return err
		}

		switch entryMetadata.EventType {

		case "event.PipelineQueue":
			// Get the run ID from the payload
			var ex event.PipelineQueue
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "queue"

		case "event.PipelineQueued":
			// Get the run ID from the payload
			var ex event.PipelineQueued
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "queued"

		case "event.PipelineLoad":
			// Get the run ID from the payload
			var ex event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "load"

		case "event.PipelineLoaded":
			// Get the run ID from the payload
			var ex event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "loaded"

		case "event.PipelineStart":
			// Get the run ID from the payload
			var ex event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "start"

		case "event.PipelineStarted":
			// Get the run ID from the payload
			var ex event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "started"

		case "event.PipelinePlan":
			// Get the run ID from the payload
			var ex event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "plan"

		case "event.PipelinePlanned":
			// Get the run ID from the payload
			var ex event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "planned"

		case "event.PipelineStepStart":
			var ex event.PipelineStepStart
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "step_start"

		case "event.PipelineStepFinished":
			var ex event.PipelineStepFinished
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "step_finished"

		case "event.PipelineFinish":
			// Get the run ID from the payload
			var ex event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "finish"

		case "event.PipelineFinished":
			// Get the run ID from the payload
			var ex event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &ex)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[ex.Event.ExecutionID] = "finished"

		default:
			// Ignore unknown types while loading
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Println(pipelineStatus)

	return nil

}

*/
