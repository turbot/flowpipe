package event

import (
	"time"
)

// Queue a mod for running in a given workspace context.
type Queue struct {
	// Host of the workspace. If empty, then assume localhost.
	CloudHost string `json:"host"`
	// The workspace context to use. May be a local workspace (e.g. default) or
	// a cloud workspace (e.g. e-gineer/scratch).
	Workspace string `json:"workspace"`
	// File system location where the mod is located, including pipeline
	// defintions.
	ModLocation string `json:"mod_location"`
	// Unique identifier for this execution.
	SpanID string `json:"span_id"`
	// Timestamp of the command
	CreatedAt time.Time `json:"created_at"`
}

// Load a mod for running in a given workspace context.
type Load struct {
	// Unique identifier for this execution.
	RunID  string `json:"run_id"`
	SpanID string `json:"span_id"`
	// Timestamp of the command
	CreatedAt time.Time `json:"created_at"`
}

type Plan Load

type Start Load

type Stop Load

type PipelineQueue struct {
	// Pipeline details
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
	// Unique identifier for this execution.
	RunID  string `json:"run_id"`
	SpanID string `json:"span_id"`
	// Timestamp of the command
	CreatedAt time.Time `json:"created_at"`
}

type PipelineLoad struct {
	RunID     string    `json:"run_id"`
	SpanID    string    `json:"span_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PipelineStart struct {
	RunID        string                 `json:"run_id"`
	SpanID       string                 `json:"span_id"`
	CreatedAt    time.Time              `json:"created_at"`
	StackID      string                 `json:"stack_id"`
	PipelineName string                 `json:"pipeline_name"`
	StepIndex    int                    `json:"step_index"`
	Input        map[string]interface{} `json:"input"`
}

type PipelinePlan PipelineStart

type PipelineFinish struct {
	RunID     string    `json:"run_id"`
	SpanID    string    `json:"span_id"`
	CreatedAt time.Time `json:"created_at"`
	StackID   string    `json:"stack_id"`
}

type PipelineStepExecute PipelineStart
