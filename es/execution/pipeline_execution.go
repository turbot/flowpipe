package execution

import "github.com/turbot/steampipe-pipelines/pipeline"

// PipelineExecution represents the execution of a single pipeline.
type PipelineExecution struct {
	// Unique identifier for this pipeline execution
	ID string `json:"id"`
	// The name of the pipeline
	Name string `json:"name"`
	// The input to the pipeline
	Args pipeline.Input `json:"args"`
	// Output from the pipeline
	Output *pipeline.Output `json:"output,omitempty"`
	// The status of the pipeline execution: queued, planned, started, completed, failed
	Status string `json:"status"`
	// Status of each step on a per-step basis. Used to determine if dependencies
	// have been met etc. Note that each step may have multiple executions, the status
	// of which are not tracked here.
	// dependencies have been met, etc. The step status is on a per-step
	StepStatus map[string]*StepStatus `json:"step_status"`
	// If this is a child pipeline, then track it's parent
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
}

// IsCanceled returns true if the pipeline has been canceled
func (pe *PipelineExecution) IsCanceled() bool {
	return pe.Status == "canceled"
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

// IsStepInitialized returns true if the step has been initialized.
func (pe *PipelineExecution) IsStepInitialized(stepName string) bool {
	return pe.StepStatus[stepName] != nil && !pe.StepStatus[stepName].Initializing
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
}

// IsComplete returns true if all executions of the step are finished or failed.
func (s *StepStatus) IsComplete() bool {
	return !s.Initializing && len(s.Queued) == 0 && len(s.Started) == 0
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
	s.Initializing = false
	s.Queued[seID] = true
	delete(s.Started, seID)
	delete(s.Finished, seID)
}

// Start marks the given execution as started.
func (s *StepStatus) Start(seID string) {
	s.Initializing = false
	delete(s.Queued, seID)
	s.Started[seID] = true
	delete(s.Finished, seID)
}

// Finish marks the given execution as finished.
func (s *StepStatus) Finish(seID string) {
	s.Initializing = false
	delete(s.Queued, seID)
	delete(s.Started, seID)
	s.Finished[seID] = true
}

// StepExecution represents the execution of a single step in a pipeline. A given
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
	Input   pipeline.Input  `json:"input"`
	ForEach *pipeline.Input `json:"for_each,omitempty"`
	// Output of the step
	Output *pipeline.Output `json:"output,omitempty"`
}
