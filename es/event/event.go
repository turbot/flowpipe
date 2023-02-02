package event

import (
	"time"
)

type runEvent struct {
	RunID     string    `json:"run_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Loaded runEvent
type Queued runEvent

type Failed struct {
	RunID        string    `json:"run_id"`
	Timestamp    time.Time `json:"timestamp"`
	ErrorMessage string    `json:"error_message"`
}

type Started struct {
	RunID        string                 `json:"run_id"`
	Timestamp    time.Time              `json:"timestamp"`
	StackID      string                 `json:"stack_id"`
	PipelineName string                 `json:"pipeline_name"`
	Input        map[string]interface{} `json:"input"`
}

type Finished struct {
	RunID     string                 `json:"run_id"`
	Timestamp time.Time              `json:"timestamp"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineStarted struct {
	RunID     string                 `json:"run_id"`
	StackID   string                 `json:"stack_id"`
	Timestamp time.Time              `json:"timestamp"`
	Input     map[string]interface{} `json:"input"`
}

type PipelinePlanned PipelineStarted

type PipelineFinished struct {
	RunID     string                 `json:"run_id"`
	StackID   string                 `json:"stack_id"`
	Timestamp time.Time              `json:"timestamp"`
	Output    map[string]interface{} `json:"output"`
}

type Executed struct {
	RunID     string                 `json:"run_id"`
	StackID   string                 `json:"stack_id"`
	Timestamp time.Time              `json:"timestamp"`
	Output    map[string]interface{} `json:"output"`
}

type ExecuteFailed struct {
	RunID        string    `json:"run_id"`
	StackID      string    `json:"stack_id"`
	Timestamp    time.Time `json:"timestamp"`
	ErrorMessage string    `json:"error_message"`
}
