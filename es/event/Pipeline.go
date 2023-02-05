package event

import (
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineQueued struct {
	Name      string                 `json:"name"`
	Input     map[string]interface{} `json:"input"`
	RunID     string                 `json:"run_id"`
	SpanID    string                 `json:"span_id"`
	CreatedAt time.Time              `json:"created_at"`
}

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
