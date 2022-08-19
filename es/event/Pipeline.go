package event

import (
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunEvent struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Timestamp     time.Time              `json:"timestamp"`
}

type PipelineRunFailed struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Timestamp     time.Time              `json:"timestamp"`
	ErrorMessage  string                 `json:"error_message"`
}

type PipelineRunQueued PipelineRunEvent

type PipelineRunLoaded struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Pipeline      pipeline.Pipeline      `json:"pipeline"`
}

type PipelineRunStarted PipelineRunLoaded

type PipelineRunStepHTTPRequestPlanned struct {
	RunID     string                 `json:"run_id"`
	Timestamp time.Time              `json:"timestamp"`
	StepID    string                 `json:"step_id"`
	Input     map[string]interface{} `json:"input"`
}

type PipelineRunStepExecuted struct {
	RunID     string                 `json:"run_id"`
	Timestamp time.Time              `json:"timestamp"`
	StepID    string                 `json:"step_id"`
	Pipeline  pipeline.Pipeline      `json:"pipeline"`
	StepIndex int                    `json:"step_index"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineRunFinished PipelineRunEvent
