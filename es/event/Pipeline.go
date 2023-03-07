package event

import (
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineQueued struct {
	Event *Event                 `json:"event"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type PipelineLoaded struct {
	Event *Event `json:"event"`
	/*
		IdentityID    string                 `json:"identity_id"`
		WorkspaceID   string                 `json:"workspace_id"`
		PipelineName  string                 `json:"pipeline_name"`
		PipelineInput map[string]interface{} `json:"pipeline_input"`
	*/
	Pipeline pipeline.Pipeline `json:"pipeline"`
}
