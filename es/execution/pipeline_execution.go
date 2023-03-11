package execution

import "github.com/turbot/steampipe-pipelines/pipeline"

// PipelineExecution represents the execution of a single pipeline.
type PipelineExecution struct {
	// Unique identifier for this pipeline execution
	ID string `json:"id"`
	// The name of the pipeline
	Name string `json:"name"`
	// The input to the pipeline
	Input map[string]interface{} `json:"input"`
	// The status of the pipeline execution: queued, planned, started, completed, failed
	Status string `json:"status"`
	// Status of each step on a per-step basis. Used to determine if dependencies
	// have been met etc. Note that each step may have multiple executions, the status
	// of which are not tracked here.
	// dependencies have been met, etc. The step status is on a per-step
	StepStatus map[string]StepStatus `json:"step_status"`
	// An ordered list of the step executions run for this pipeline. Details
	// of each step execution are available in the StepExecutions map on the Execution.
	StepExecutions []string `json:"step_executions"`
	// If this is a child pipeline, then track it's parent
	ParentStepExecutionID string `json:"parent_step_execution_id,omitempty"`
}

// StepStatus represents the status of a single step in a pipeline. It reflects the
// status of the step as a whole, not the status of each execution of the step.
type StepStatus struct {
	Queued   int `json:"queued"`
	Started  int `json:"started"`
	Finished int `json:"finished"`
	Failed   int `json:"failed"`
}

// Progress returns a percentage complete for the executions of the step.
// It always returns an integer.
func (s StepStatus) Progress() int {
	total := s.Queued + s.Started + s.Finished + s.Failed
	if total == 0 {
		return 0
	}
	return (s.Finished + s.Failed) * 100 / total
}

func (s StepStatus) Total() int {
	return s.Queued + s.Started + s.Finished + s.Failed
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
	// Output of the step
	Output pipeline.StepOutput `json:"output"`
}
