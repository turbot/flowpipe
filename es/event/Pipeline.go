package event

import (
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunEvent struct {
	Name      string                 `json:"name"`
	Input     map[string]interface{} `json:"input"`
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
}

type PipelineRunFailed struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	SpanID        string                 `json:"span_id"`
	CreatedAt     time.Time              `json:"created_at"`
	ErrorMessage  string                 `json:"error_message"`
}

type PipelineQueued PipelineRunEvent

type PipelineLoaded struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	SpanID        string                 `json:"span_id"`
	CreatedAt     time.Time              `json:"created_at"`
	Pipeline      pipeline.Pipeline      `json:"pipeline"`
}

type PipelineRunStarted PipelineLoaded

type PipelineRunStepPrimitivePlanned struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
	StepID    string                 `json:"step_id"`
	Primitive string                 `json:"primitive"`
	Input     map[string]interface{} `json:"input"`
}

type PipelineRunStepHTTPRequestPlanned struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
	StepID    string                 `json:"step_id"`
	Input     map[string]interface{} `json:"input"`
}

type PipelineRunStepExecuted struct {
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
	StepID    string                 `json:"step_id"`
	Pipeline  pipeline.Pipeline      `json:"pipeline"`
	StepIndex int                    `json:"step_index"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineRunStepFailed struct {
	RunID        string            `json:"run_id"`
	SpanID       string            `json:"span_id"`
	CreatedAt    time.Time         `json:"created_at"`
	StepID       string            `json:"step_id"`
	Pipeline     pipeline.Pipeline `json:"pipeline"`
	ErrorMessage string            `json:"error_message"`
}

type PipelineRunFinished PipelineRunEvent
