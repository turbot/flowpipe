package event

type Queue struct {
	IdentityID   string                 `json:"identity_id"`
	WorkspaceID  string                 `json:"workspace_id"`
	PipelineName string                 `json:"pipeline_name"`
	Input        map[string]interface{} `json:"input"`
	RunID        string                 `json:"run_id"`
}

type PipelineStart struct {
	RunID        string                 `json:"run_id"`
	StackID      string                 `json:"stack_id"`
	PipelineName string                 `json:"pipeline_name"`
	StepIndex    int                    `json:"step_index"`
	Input        map[string]interface{} `json:"input"`
}

type PipelineFinish struct {
	RunID   string `json:"run_id"`
	StackID string `json:"stack_id"`
}

type Execute PipelineStart
