package execution

import (
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
)

// PipelineExecution represents the execution of a single types.
type PipelineExecution struct {
	// Unique identifier for this pipeline execution
	ID string `json:"id"`
	// The name of the pipeline
	Name string `json:"name"`
	// The input to the pipeline
	Args types.Input `json:"args"`
	// Output from the pipeline
	Output *types.StepOutput `json:"output,omitempty"`
	// The status of the pipeline execution: queued, planned, started, completed, failed
	Status string `json:"status"`
	// Status of each step on a per-step basis. Used to determine if dependencies
	// have been met etc. Note that each step may have multiple executions, the status
	// of which are not tracked here.
	// dependencies have been met, etc. The step status is on a per-step
	StepStatus map[string]*StepStatus `json:"step_status"`

	// If this is a child pipeline, then track it's parent
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`

	Errors map[string]types.StepError `json:"errors,omitempty"`
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
func (pe *PipelineExecution) IsStepFinalFailure(step types.IPipelineHclStep, ex *Execution) bool {
	if !pe.IsStepFail(step.GetName()) {
		// Step not failed, so no need to calculate, return false
		return false
	}

	var failedStepExecutions []StepExecution
	if step.GetError().Retries > 0 && !step.GetError().Ignore {
		if pe.StepStatus[step.GetName()].FailCount() > step.GetError().Retries {
			failedStepExecutions = ex.PipelineStepExecutions(pe.ID, step.GetName())

			if failedStepExecutions[len(failedStepExecutions)-1].Error == nil {
				pe.Fail(step.GetName(), types.StepError{Detail: fperr.InternalWithMessage("change this pipeline error - THERE IS SOMETHING WRONG HERE?")})
			} else {
				// Set the error
				pe.Fail(step.GetName(), *failedStepExecutions[len(failedStepExecutions)-1].Error)
			}
			// pe.Fail(step.GetName(), types.StepError{Detail: fperr.InternalWithMessage("change this pipeline error")})
			return true
		} else {
			return false
		}
	} else if !step.GetError().Ignore {
		failedStepExecutions = ex.PipelineStepExecutions(pe.ID, step.GetName())
		pe.Fail(step.GetName(), *failedStepExecutions[len(failedStepExecutions)-1].Error)
		return true
	}
	return true

}
func (pe *PipelineExecution) Fail(stepName string, stepError types.StepError) {
	pe.Errors[stepName] = stepError
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
func (pe *PipelineExecution) QueueStep(stepName string, seID string) {
	pe.StepStatus[stepName].Queue(seID)
}

// StartStep marks the given step execution as started.
func (pe *PipelineExecution) StartStep(stepName string, seID string) {
	pe.StepStatus[stepName].Start(seID)
}

// FinishStep marks the given step execution as started.
func (pe *PipelineExecution) FinishStep(stepName string, seID string) {
	pe.StepStatus[stepName].Finish(seID)
}

func (pe *PipelineExecution) FailStep(stepName string, seID string) {
	pe.StepStatus[stepName].Fail(seID)
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
		panic(fperr.BadRequestWithMessage("Step " + seID + " already failed"))
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
		panic(fperr.BadRequestWithMessage("Step " + seID + " already failed"))
	}

	s.Initializing = false
	delete(s.Queued, seID)
	s.Started[seID] = true
}

// Finish marks the given execution as finished.
func (s *StepStatus) Finish(seID string) {
	// Can't finish if the step already set to fail (safety check)
	if s.Failed[seID] {
		panic(fperr.BadRequestWithMessage("Step " + seID + " already failed"))
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
		panic(fperr.BadRequestWithMessage("Step " + seID + " already failed"))
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
	// The status of the step execution: queued, planned, started, completed, failed
	Status string `json:"status"`
	// Input to the step
	Input   types.Input  `json:"input"`
	ForEach *types.Input `json:"for_each,omitempty"`
	// Output of the step
	Output *types.StepOutput `json:"output,omitempty"`

	// TODO: should we just put fperr.ErrorModel here?
	Error *types.StepError `json:"error,omitempty"`
}
