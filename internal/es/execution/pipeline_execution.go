package execution

import (
	"github.com/turbot/flowpipe/pipeparser/pcerr"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/zclconf/go-cty/cty"
)

// PipelineExecution represents the execution of a single types.
type PipelineExecution struct {
	// Unique identifier for this pipeline execution
	ID string `json:"id"`
	// The name of the pipeline
	Name string `json:"name"`
	// The input to the pipeline
	Args pipeline.Input `json:"args,omitempty"`

	// The output of the pipeline
	PipelineOutput map[string]interface{} `json:"pipeline_output,omitempty"`

	// The status of the pipeline execution: queued, planned, started, completed, failed
	Status string `json:"status"`

	// Status of each step on a per-step basis. Used to determine if dependencies
	// have been met etc. Note that each step may have multiple executions, the status
	// of which are not tracked here.
	// dependencies have been met, etc. The step status is on a per-step
	StepStatus map[string]*StepStatus `json:"-"`

	// If this is a child pipeline, then track it's parent
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
	ParentExecutionID     string `json:"parent_execution_id,omitempty"`

	// All errors from the step execution + any errors that can be added to the pipeline execution manually
	Errors []pipeline.StepError `json:"errors,omitempty"`

	// The "final" output for all the steps in this pipeline execution.
	AllStepOutputs ExecutionStepOutputs `json:"-"`

	// Steps triggered by pipelines in the execution.
	StepExecutions map[string]*StepExecution `json:"step_executions,omitempty"`

	// TODO: not sure if we need this, it's a different index of the step executions
	// TODO: but also a way to track the order of execution for a given step
	StepExecutionOrder map[string][]string `json:"-"`
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

	for stepType, v := range pe.AllStepOutputs {

		if stepVariables[stepType] == cty.NilVal {
			stepVariables[stepType] = cty.ObjectVal(map[string]cty.Value{})
		}

		vm := stepVariables[stepType].AsValueMap()
		if vm == nil {
			vm = map[string]cty.Value{}
		}

		for stepName, stepOutput := range v {
			if nonIndexStepOutput, ok := stepOutput.(*pipeline.Output); ok {
				var err error
				vm[stepName], err = nonIndexStepOutput.AsCtyValue()
				if err != nil {
					return nil, err
				}
			} else if indexedStepOutput, ok := stepOutput.([]*pipeline.Output); ok {
				var err error

				ctyValList := make([]cty.Value, len(indexedStepOutput))
				for i, stepOutput := range indexedStepOutput {
					ctyValList[i], err = stepOutput.AsCtyValue()
					if err != nil {
						return nil, err
					}
				}
				vm[stepName] = cty.TupleVal(ctyValList)
			}
		}

		stepVariables[stepType] = cty.ObjectVal(vm)
	}

	executionVariables := map[string]cty.Value{
		schema.BlockTypePipelineStep: cty.ObjectVal(stepVariables),
	}

	return executionVariables, nil
}

// PipelineStepExecutions returns a list of step executions for the given
// pipeline execution ID and step name.
func (pe *PipelineExecution) OrderedStepExecutions(stepName string) []StepExecution {

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

// IsComplete returns true if all steps (that have been initialized) are complete.
func (pe *PipelineExecution) IsComplete() bool {
	complete := true
	for _, status := range pe.StepStatus {
		if !status.IsComplete() {
			complete = false
			break
		}
	}
	return complete
}

// IsStepComplete returns true if all executions of the step are finished.
func (pe *PipelineExecution) IsStepComplete(stepName string) bool {
	return pe.StepStatus[stepName] != nil && pe.StepStatus[stepName].IsComplete()
}

func (pe *PipelineExecution) IsStepFail(stepName string) bool {
	return pe.StepStatus[stepName] != nil && pe.StepStatus[stepName].IsFail()
}

// Calculate if this step needs to be retried, or this is the final failure of the step
func (pe *PipelineExecution) IsStepFinalFailure(step pipeline.IPipelineStep, ex *Execution) bool {

	return true
	// if !pe.IsStepFail(step.GetFullyQualifiedName()) {
	// 	// Step not failed, so no need to calculate, return false
	// 	return false
	// }

	// var failedStepExecutions []StepExecution
	// if step.GetError().Retries > 0 && !step.GetError().Ignore {
	// 	if pe.StepStatus[step.GetFullyQualifiedName()].FailCount() > step.GetError().Retries {
	// 		failedStepExecutions = ex.PipelineStepExecutions(pe.ID, step.GetFullyQualifiedName())

	// 		if failedStepExecutions[len(failedStepExecutions)-1].Error == nil {
	// 			pe.Fail(step.GetFullyQualifiedName(), pipeline.StepError{Detail: fperr.InternalWithMessage("change this pipeline error - THERE IS SOMETHING WRONG HERE?")})
	// 		} else {
	// 			// Set the error
	// 			pe.Fail(step.GetFullyQualifiedName(), *failedStepExecutions[len(failedStepExecutions)-1].Error)
	// 		}
	// 		// pe.Fail(step.GetName(), pipeline.StepError{Detail: fperr.InternalWithMessage("change this pipeline error")})
	// 		return true
	// 	} else {
	// 		return false
	// 	}
	// } else if !step.GetError().Ignore {
	// 	failedStepExecutions = ex.PipelineStepExecutions(pe.ID, step.GetFullyQualifiedName())
	// 	pe.Fail(step.GetFullyQualifiedName(), *failedStepExecutions[len(failedStepExecutions)-1].Error)
	// 	return true
	// }
	// return true

}
func (pe *PipelineExecution) Fail(stepName string, stepError ...pipeline.StepError) {
	pe.Errors = append(pe.Errors, stepError...)
}

// IsStepInitialized returns true if the step has been initialized.
func (pe *PipelineExecution) IsStepInitialized(stepName string) bool {
	return pe.StepStatus[stepName] != nil && !pe.StepStatus[stepName].Initializing
}

// TODO: this doesn't work for step execution retry, it assumes that the entire step
// TODO: must be retried
func (pe *PipelineExecution) IsStepQueued(stepName string) bool {
	return pe.StepStatus[stepName] != nil && len(pe.StepStatus[stepName].Queued) > 0
}

// InitializeStep initializes the step status for the given step.
func (pe *PipelineExecution) InitializeStep(stepName string) {
	if pe.StepStatus[stepName] != nil {
		// Step is already initialized
		return
	}
	pe.StepStatus[stepName] = &StepStatus{
		Initializing: true,
		Queued:       map[string]bool{},
		Started:      map[string]bool{},
		Finished:     map[string]bool{},
		Failed:       map[string]bool{},
	}
}

// QueueStep marks the given step execution as queued.
func (pe *PipelineExecution) QueueStep(stepFullyQualifiedName string, seID string) {
	pe.StepStatus[stepFullyQualifiedName].Queue(seID)
}

// StartStep marks the given step execution as started.
func (pe *PipelineExecution) StartStep(stepFullyQualifiedName string, seID string) {
	pe.StepStatus[stepFullyQualifiedName].Start(seID)
}

// FinishStep marks the given step execution as started.
func (pe *PipelineExecution) FinishStep(stepFullyQualifiedName string, seID string) {
	pe.StepStatus[stepFullyQualifiedName].Finish(seID)
}

func (pe *PipelineExecution) FailStep(stepFullyQualifiedName string, seID string) {
	pe.StepStatus[stepFullyQualifiedName].Fail(seID)
}

// This needs to be a map because if we have a for loop, each loop will have a different step execution id
type StepStatus struct {
	// When the step is initializing it doesn't yet have any executions.
	// We track it as initializing until the first execution is queued.
	Initializing bool `json:"initializing"`
	// Step executions that are queued.
	Queued map[string]bool `json:"queued"`
	// Step executions that are started.
	Started map[string]bool `json:"started"`
	// Step executions that are finished.
	Finished map[string]bool `json:"finished"`
	// Step executions that are failed.
	Failed map[string]bool `json:"failed"`
}

// IsComplete returns true if all executions of the step are finished or failed.
func (s *StepStatus) IsComplete() bool {
	// One step can have more than 1 execution, for example if a step has a for_each directive
	// or retries
	return !s.Initializing && len(s.Queued) == 0 && len(s.Started) == 0
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
		panic(pcerr.BadRequestWithMessage("Step " + seID + " already failed"))
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
		panic(pcerr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false
	delete(s.Queued, seID)
	s.Started[seID] = true
}

// Finish marks the given execution as finished.
func (s *StepStatus) Finish(seID string) {
	// Can't finish if the step already set to fail (safety check)
	if s.Failed[seID] {
		panic(pcerr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false

	// Important to delete queued and started so we know that the step has "completed"
	delete(s.Queued, seID)
	delete(s.Started, seID)
	s.Finished[seID] = true
}

func (s *StepStatus) Fail(seID string) {
	// Can't fail if the step already finished (safety check)
	if s.Finished[seID] {
		panic(pcerr.BadRequestWithMessage("Step " + seID + " already failed"))
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
	Input pipeline.Input `json:"input"`

	// for_each controls
	StepForEach *pipeline.StepForEach `json:"step_for_each,omitempty"`

	NextStepAction pipeline.NextStepAction `json:"next_step_action,omitempty"`

	// Output of the step
	Output *pipeline.Output `json:"output,omitempty"`
}

func (se *StepExecution) Index() *int {
	if se.StepForEach == nil {
		return nil
	}

	return &se.StepForEach.Index
}
