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

type PipelinePlanned struct {
	RunID         string                 `json:"run_id"`
	SpanID        string                 `json:"span_id"`
	StackID       string                 `json:"stack_id"`
	NextStepIndex int                    `json:"next_step_index"`
	CreatedAt     time.Time              `json:"created_at"`
	Input         map[string]interface{} `json:"input"`
}

type PipelineFinished struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	StackID   string                 `json:"stack_id"`
	CreatedAt time.Time              `json:"created_at"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineStepExecuted struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	StackID   string                 `json:"stack_id"`
	StepIndex int                    `json:"step_index"`
	CreatedAt time.Time              `json:"created_at"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineFailed struct {
	RunID        string    `json:"run_id"`
	SpanID       string    `json:"span_id"`
	StackID      string    `json:"stack_id"`
	CreatedAt    time.Time `json:"created_at"`
	ErrorMessage string    `json:"error_message"`
}
