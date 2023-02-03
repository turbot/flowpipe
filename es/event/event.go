package event

import (
	"time"
)

type runEvent struct {
	RunID     string    `json:"run_id"`
	SpanID    string    `json:"span_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Loaded runEvent
type Queued runEvent

type Failed struct {
	RunID        string    `json:"run_id"`
	SpanID       string    `json:"span_id"`
	CreatedAt    time.Time `json:"created_at"`
	ErrorMessage string    `json:"error_message"`
}

type Started struct {
	RunID        string                 `json:"run_id"`
	SpanID       string                 `json:"span_id"`
	CreatedAt    time.Time              `json:"created_at"`
	StackID      string                 `json:"stack_id"`
	PipelineName string                 `json:"pipeline_name"`
	Input        map[string]interface{} `json:"input"`
}

type Planned Started

type Stopped struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineStarted struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	StackID   string                 `json:"stack_id"`
	CreatedAt time.Time              `json:"created_at"`
	Input     map[string]interface{} `json:"input"`
}

type PipelinePlanned PipelineStarted

type PipelineFinished struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	StackID   string                 `json:"stack_id"`
	CreatedAt time.Time              `json:"created_at"`
	Output    map[string]interface{} `json:"output"`
}

type Executed struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	StackID   string                 `json:"stack_id"`
	CreatedAt time.Time              `json:"created_at"`
	Output    map[string]interface{} `json:"output"`
}

type ExecuteFailed struct {
	RunID        string    `json:"run_id"`
	SpanID       string    `json:"span_id"`
	StackID      string    `json:"stack_id"`
	CreatedAt    time.Time `json:"created_at"`
	ErrorMessage string    `json:"error_message"`
}
