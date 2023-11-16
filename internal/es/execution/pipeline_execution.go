package execution

import (
	"strconv"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

// PipelineExecution represents the execution of a single types.
type PipelineExecution struct {
	// Unique identifier for this pipeline execution
	ID string `json:"id"`
	// The name of the pipeline
	Name string `json:"name"`
	// The input to the pipeline
	Args modconfig.Input `json:"args,omitempty"`

	// The output of the pipeline
	PipelineOutput map[string]interface{} `json:"pipeline_output,omitempty"`

	// The status of the pipeline execution: queued, planned, started, completed, failed
	Status string `json:"status"`

	// Status of each step on a per-step index basis. Used to determine if dependencies
	// have been met etc. Note that each step may have multiple executions, the status
	// of which are not tracked here.
	// dependencies have been met, etc.
	//
	// The Step Status used to be per-step, however the addition of for_each means that we now need to expand this
	// tracking to include the "index" of the step
	//
	// for_each have 2 type of results: list or map, however in Flowpipe they are both treated as a map,
	// the list is simply a map that the key happens to be a string of "0", "1", "2"
	//
	/*
		The data structure of StepStatus is as follow:
		{
			"echo.echo": {
				"0": {
					xyz
				},
				"1": {
					xyz
				}
			},
			"http.one": {
				"foo": {
					zzz
				},
				"bar": {
					yyy
				}
			}
		}

		echo.echo has a for_each which is a list, so the key is the index of the list

		http.one has a for_each which is a map, so the key is the key of the map

		LOOP

		Loop will be recorded in StepStatus.StepExecution, it's an array
		**/
	StepStatus map[string]map[string]*StepStatus `json:"step_status,omitempty"`

	// If this is a child pipeline, then track it's parent
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
	ParentExecutionID     string `json:"parent_execution_id,omitempty"`

	// All errors from the step execution + any errors that can be added to the pipeline execution manually
	Errors []modconfig.StepError `json:"errors,omitempty"`

	// Steps triggered by pipelines in the execution.
	StepExecutions map[string]*StepExecution `json:"-"`

	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

/*
*

	Arrange the step outputs in a way that it can be used for HCL Expression evaluation

	The expressions look something like: step.echo.text_1.text

	So we need to arrange the output as such:

	"step": {
		"echo": {
			"text_1": {
				"text": "hello world" <-- this is the output from the step
			},
			"text_2": {
				"text": "hello world" <-- this is the output from the step
			},
		},
		"http": {
			"my_http": {
				"response_body": "hello world" <-- this is the output from the step
			},
		},
	},
	"param": {
		"my_param": "hello world" <-- this is set by the calling function, but maybe we should do it here?
	}
*/
func (pe *PipelineExecution) GetExecutionVariables() (map[string]cty.Value, error) {
	stepVariables := make(map[string]cty.Value)

	pipelineDefn, err := db.GetPipeline(pe.Name)

	if err != nil {
		return nil, err
	}

	for stepFullName, stepStatus := range pe.StepStatus {
		parts := strings.Split(stepFullName, ".")

		if len(parts) != 2 {
			return nil, perr.InternalWithMessage("Invalid step full name: " + stepFullName + " it needs to be in the <step type>.<step name> format.")
		}

		stepDefn := pipelineDefn.GetStep(stepFullName)
		stepType := parts[0]
		stepName := parts[1]

		if stepVariables[stepType].IsNull() {
			stepVariables[stepType] = cty.ObjectVal(map[string]cty.Value{})
		}

		stepTypeValueMap := stepVariables[stepType].AsValueMap()
		if stepTypeValueMap == nil {
			stepTypeValueMap = map[string]cty.Value{}
		}

		if stepTypeValueMap[stepName].IsNull() {
			stepTypeValueMap[stepName] = cty.ObjectVal(map[string]cty.Value{})
		}

		forEach := stepDefn.GetForEach() != nil
		loop := stepDefn.GetUnresolvedBodies()["loop"] != nil

		if !forEach {
			if len(stepStatus) > 1 {
				return nil, perr.InternalWithMessage("Step " + stepFullName + " has more than element in StepStatus. This is unexpected, for a step that does not have for_each there should never be more than 1 element in the StepStatus ")
			}
			singleStepStatus := stepStatus["0"]

			if singleStepStatus == nil || len(singleStepStatus.StepExecutions) == 0 {
				continue
			}

			// Retry will have multiple step executions per step instance, however we should still structure the data as if
			// there is no loop
			//
			// Similar vein with loop + error retry
			//
			// Say we have 3 loops, but the middle one is retried twice we will have 5 step executions:
			// 0, 1, 1, 1, 2
			//
			// But the index in the EvalContext should still be "0", "1", "2"

			singleStepValueMap, err := buildSingleStepStatusOutput(stepName, loop, singleStepStatus)

			if err != nil {
				return nil, err
			}
			stepTypeValueMap[stepName] = cty.ObjectVal(singleStepValueMap)
		} else {
			indexedStepNameValueMap := stepTypeValueMap[stepName].AsValueMap()
			if indexedStepNameValueMap == nil {
				indexedStepNameValueMap = map[string]cty.Value{}
			}

			// there is for_each, we need to loop through all the instance of the step and then the step execution for each of that instance
			for k, singleStepStatus := range stepStatus {
				if singleStepStatus == nil || len(singleStepStatus.StepExecutions) == 0 {
					continue
				}
				singleStepValueMap, err := buildSingleStepStatusOutput(stepName, loop, singleStepStatus)

				if err != nil {
					return nil, err
				}
				indexedStepNameValueMap[k] = cty.ObjectVal(singleStepValueMap)
			}
			stepTypeValueMap[stepName] = cty.ObjectVal(indexedStepNameValueMap)
		}
		stepVariables[stepType] = cty.ObjectVal(stepTypeValueMap)
	}

	executionVariables := map[string]cty.Value{
		schema.BlockTypePipelineStep: cty.ObjectVal(stepVariables),
	}

	return executionVariables, nil
}

func buildSingleStepStatusOutput(stepName string, loop bool, singleStepStatus *StepStatus) (map[string]cty.Value, error) {

	// TODO: error
	// TODO: retry

	var err error
	var stepNameValueMap map[string]cty.Value
	if !loop {
		if len(singleStepStatus.StepExecutions) == 0 {
			return map[string]cty.Value{}, nil
		}

		// Get the last step executions
		lastStepExecution := singleStepStatus.StepExecutions[len(singleStepStatus.StepExecutions)-1]
		stepNameValueMap, err = BuildSingleStepExecutionOutput(&lastStepExecution, stepName)
		if err != nil {
			return nil, err
		}

	} else {
		if len(singleStepStatus.StepExecutions) == 0 {
			return map[string]cty.Value{}, nil
		}

		stepNameValueMap = map[string]cty.Value{}

		// for each step execution, we need to get the output and add it to the map
		for index := range singleStepStatus.StepExecutions {
			indexedStepValueMap, err := BuildSingleStepExecutionOutput(&singleStepStatus.StepExecutions[index], stepName)
			if err != nil {
				return nil, err
			}

			// don't use stepExecution.StepLoop.Key because it's intended for HCL expression evaluation
			// not for arranging the output, the index will always be 1 ahead
			key := strconv.Itoa(index)
			stepNameValueMap[key] = cty.ObjectVal(indexedStepValueMap)
		}
	}

	return stepNameValueMap, nil
}

func BuildSingleStepExecutionOutput(lastStepExecution *StepExecution, stepName string) (map[string]cty.Value, error) {
	var singleStepValueMap map[string]cty.Value
	var err error

	if lastStepExecution.Output != nil {
		if lastStepExecution.Output.Status == "skipped" {
			return singleStepValueMap, nil
		}
		singleStepValueMap, err = lastStepExecution.Output.AsCtyMap()
		if err != nil {
			return nil, err
		}
	}

	// do we have configured step output? this is the output block on the step
	if len(lastStepExecution.StepOutput) > 0 {
		if singleStepValueMap["output"].IsNull() {
			singleStepValueMap["output"], err = hclhelpers.ConvertMapToCtyValue(lastStepExecution.StepOutput)
			if err != nil {
				return nil, perr.InternalWithMessage("Unable to convert map to cty value " + err.Error())
			}
		} else {
			stepOutput := singleStepValueMap["output"].AsValueMap()

			for configuredOutputName, configuredOutputValue := range lastStepExecution.StepOutput {
				if !stepOutput[configuredOutputName].IsNull() {
					return nil, perr.BadRequestWithMessage("output block '" + configuredOutputName + "' already exists in step '" + stepName + "'")
				}
				stepOutput[configuredOutputName], err = hclhelpers.ConvertInterfaceToCtyValue(configuredOutputValue)
				if err != nil {
					return nil, perr.InternalWithMessage("Unable to convert interface to cty value " + err.Error())
				}
			}
			singleStepValueMap["output"] = cty.ObjectVal(stepOutput)
		}
	}

	return singleStepValueMap, nil
}

// IsCanceled returns true if the pipeline has been canceled
func (pe *PipelineExecution) IsCanceled() bool {
	return pe.Status == "canceled"
}

// IsPaused returns true if the pipeline has been paused
func (pe *PipelineExecution) IsPaused() bool {
	return pe.Status == "paused"
}

func (pe *PipelineExecution) IsFail() bool {
	return pe.Status == "failed"
}

func (pe *PipelineExecution) IsFinished() bool {
	return pe.Status == "finished"
}

func (pe *PipelineExecution) IsFinishing() bool {
	return pe.Status == "finishing"
}

func (pe *PipelineExecution) ShouldFail() bool {
	return len(pe.Errors) > 0

}

// IsComplete returns true if all steps are complete.
func (pe *PipelineExecution) IsComplete() bool {
	pipeline, err := db.GetPipeline(pe.Name)
	if err != nil {
		// TODO: what do we do here?
		return false
	}

	if len(pe.StepStatus) != len(pipeline.Steps) {
		return false
	}

	for _, indexedStatus := range pe.StepStatus {
		// If indexedStatus is nil, then the step hasn't been initialized
		// TODO: for_each - this concept of step initialization does not work well with for_each when each instance of for_each has a loop
		if len(indexedStatus) == 0 {
			return false
		}

		for _, status := range indexedStatus {
			if !status.IsComplete() {
				return false
			}
		}
	}
	return true
}

// IsStepComplete returns true if all executions of the step are finished.
func (pe *PipelineExecution) IsStepComplete(stepName string) bool {
	if pe.StepStatus[stepName] == nil || len(pe.StepStatus[stepName]) == 0 {
		return false
	}

	for _, s := range pe.StepStatus[stepName] {
		if !s.IsComplete() {
			return false
		}
	}
	return true
}

func (pe *PipelineExecution) IsStepFail(stepName string) bool {
	if pe.StepStatus[stepName] == nil || len(pe.StepStatus[stepName]) == 0 {
		return false
	}

	for _, s := range pe.StepStatus[stepName] {
		if !s.IsFail() {
			return false
		}
	}
	return true
}

// Calculate if this step needs to be retried, or this is the final failure of the step
func (pe *PipelineExecution) IsStepFinalFailure(step modconfig.PipelineStep, ex *Execution) bool {
	return true
}

// TODO: this is where we collect the failures so the "ShouldFail" test works .. not sure if this is the correct place?
func (pe *PipelineExecution) Fail(stepName string, stepError ...modconfig.StepError) {
	pe.Errors = append(pe.Errors, stepError...)
}

// IsStepInitialized returns true if the step has been initialized.
func (pe *PipelineExecution) IsStepInitialized(stepName string) bool {
	if pe.StepStatus[stepName] == nil || len(pe.StepStatus[stepName]) == 0 {
		return false
	}

	// for _, s := range pe.StepStatus[stepName] {
	// 	if !s.Initializing {
	// 		return false
	// 	}
	// }
	return true
}

func (pe *PipelineExecution) IsStepInLoopHold(stepName string) bool {
	return false
}

// TODO: this doesn't work for step execution retry, it assumes that the entire step
// TODO: must be retried
func (pe *PipelineExecution) IsStepQueued(stepName string) bool {
	if pe.StepStatus[stepName] == nil || len(pe.StepStatus[stepName]) == 0 {
		return false
	}

	for _, s := range pe.StepStatus[stepName] {
		if len(s.Queued) > 0 {
			return true
		}
	}
	return false
}

// InitializeStep initializes the step status for the given step.
func (pe *PipelineExecution) InitializeStep(stepName string) {
	if pe.StepStatus[stepName] != nil {
		return
	}
	pe.StepStatus[stepName] = map[string]*StepStatus{}
}

// QueueStep marks the given step execution as queued.
func (pe *PipelineExecution) QueueStep(stepFullyQualifiedName, key, seID string) {

	if pe.StepStatus[stepFullyQualifiedName][key] == nil {
		pe.StepStatus[stepFullyQualifiedName][key] = &StepStatus{
			Initializing: true,
			Queued:       map[string]bool{},
			Started:      map[string]bool{},
			Finished:     map[string]bool{},
			Failed:       map[string]bool{},
		}
	}

	pe.StepStatus[stepFullyQualifiedName][key].Queue(seID)
}

// StartStep marks the given step execution as started.
func (pe *PipelineExecution) StartStep(stepFullyQualifiedName, key, seID string) {
	pe.StepStatus[stepFullyQualifiedName][key].Start(seID)
}

// FinishStep marks the given step execution as started.
func (pe *PipelineExecution) FinishStep(stepFullyQualifiedName, key, seID string, loopHold, errorHold bool) {
	pe.StepStatus[stepFullyQualifiedName][key].Finish(seID, loopHold, errorHold)
}

func (pe *PipelineExecution) FailStep(stepFullyQualifiedName, key, seID string) {
	pe.StepStatus[stepFullyQualifiedName][key].Fail(seID)
}

// This needs to be a map because if we have a for loop, each loop will have a different step execution id
type StepStatus struct {
	// When the step is initializing it doesn't yet have any executions.
	// We track it as initializing until the first execution is queued.
	Initializing bool   `json:"initializing"`
	OverralState string `json:"overral_state"`

	//
	// Both LoopHold and ErrorHold must be resolved **before** the "finish" event is called, i.e. it needs to be calculated at the
	// end of "step start command" and "step pipeline finish" command.
	//
	// It can't be calculated at the "finish" event because it's already too late. If the planner see that it has an finish
	// event without either a LoopHold or ErrorHold, it will mark the step as completed or failed
	//
	// Indicates that step is in a loop so we don't mark it as finished
	LoopHold bool `json:"loop_hold"`

	// Indicates that a step is in retry loop so we don't mark it as failed
	ErrorHold bool `json:"error_hold"`

	// Step executions that are queued.
	Queued map[string]bool `json:"queued"`
	// Step executions that are started.
	Started map[string]bool `json:"started"`
	// Step executions that are finished.
	Finished map[string]bool `json:"finished"`
	// Step executions that are failed.
	Failed map[string]bool `json:"failed"`

	// There's the step execution in execution, this is the same but in a list for a given step status
	// The element in this slice should point to the same element in the StepExecutions map (in PipelineExecution)
	StepExecutions []StepExecution `json:"step_executions"`
}

// IsComplete returns true if all executions of the step are finished or failed.
func (s *StepStatus) IsComplete() bool {
	if s == nil {
		return false
	}
	// One step can have more than 1 execution, for example if a step has a for_each directive
	// or retries
	if s.OverralState == "empty_for_each" {
		return true
	}
	return !s.Initializing && len(s.Queued) == 0 && len(s.Started) == 0 && !s.LoopHold && !s.ErrorHold
}

func (s *StepStatus) IsStarted() bool {
	if s == nil {
		return false
	}
	return s.Initializing || len(s.Queued) > 0 || len(s.Started) > 0 || !s.LoopHold || !s.ErrorHold
}

// IsFail returns true if any executions of the step failed.
func (s *StepStatus) IsFail() bool {
	return len(s.Failed) > 0
}

func (s *StepStatus) FailCount() int {
	return len(s.Failed)
}

// Progress returns the percentage of executions of the step that are complete.
func (s *StepStatus) Progress() int {
	if s.Initializing {
		return 0
	}
	total := len(s.Queued) + len(s.Started) + len(s.Finished)
	if total == 0 {
		return 0
	}
	return len(s.Finished) * 100 / total
}

// Queue marks the given execution as queued.
func (s *StepStatus) Queue(seID string) {
	// Can't queue if the step already finished or started (safety check)
	if s.Finished[seID] || s.Failed[seID] {
		panic(perr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false
	s.Queued[seID] = true
	delete(s.Started, seID)
	delete(s.Finished, seID)
}

// Start marks the given execution as started.
func (s *StepStatus) Start(seID string) {
	// Can't start if the step already finished or started (safety check)
	if s.Finished[seID] || s.Failed[seID] {
		panic(perr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false
	delete(s.Queued, seID)
	s.Started[seID] = true
}

// Finish marks the given execution as finished.
func (s *StepStatus) Finish(seID string, loopHold, errorHold bool) {
	// Can't finish if the step already set to fail (safety check)
	if s.Failed[seID] {
		panic(perr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.LoopHold = loopHold
	s.ErrorHold = errorHold

	s.Initializing = false

	// Important to delete queued and started so we know that the step has "completed"
	delete(s.Queued, seID)
	delete(s.Started, seID)
	s.Finished[seID] = true
}

func (s *StepStatus) Fail(seID string) {
	// Can't fail if the step already finished (safety check)
	if s.Finished[seID] {
		panic(perr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false

	// Important to delete queued and started so we know that the step has "completed"
	delete(s.Queued, seID)
	delete(s.Started, seID)
	s.Failed[seID] = true
}

// StepExecution represents the execution of a single step in a types. A given
// step definition may be executed multiple times.
type StepExecution struct {
	// Unique identifier for this step execution
	PipelineExecutionID string `json:"pipeline_execution_id"`
	ID                  string `json:"id"`

	// The name of the step in the pipeline definition
	Name string `json:"name"`

	// The status of the step execution: "started", "finished", "failed", "skipped"
	Status string `json:"status"`

	// Input to the step
	Input modconfig.Input `json:"input"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *modconfig.StepLoop    `json:"step_loop,omitempty"`
	StepRetry   *modconfig.StepRetry   `json:"step_retry,omitempty"`

	NextStepAction modconfig.NextStepAction `json:"next_step_action,omitempty"`

	// Native/primitive output of the step
	Output *modconfig.Output `json:"output,omitempty"`

	// The output from the Step's output block:
	// output "foo" {
	//    value = <xxx>
	//	}
	//
	StepOutput map[string]interface{} `json:"step_output,omitempty"`

	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

func (se *StepExecution) Key() *string {
	if se.StepForEach == nil {
		return nil
	}

	return &se.StepForEach.Key
}
