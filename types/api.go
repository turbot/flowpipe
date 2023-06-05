package types

// APIVersionRequestURI defines the requested API version.
type APIVersionRequestURI struct {
	APIVersion string `uri:"api_version" binding:"required,flowpipe_api_version"`
}

type ListRequestQuery struct {
	NextToken string `json:"next_token" form:"next_token" binding:"omitempty"`
	Limit     *int   `json:"limit,omitempty" form:"limit" binding:"omitempty"`
}

type PipelineRequestURI struct {
	PipelineName string `uri:"pipeline_name" binding:"required" format:"^[a-z_]{0,32}$"`
}

type TriggerRequestURI struct {
	TriggerName string `uri:"trigger_name" binding:"required" format:"^[a-z]{0,32}$"`
}

type VariableRequestURI struct {
	VariableName string `uri:"variable_name" binding:"required" format:"^[a-z]{0,32}$"`
}

type ProcessRequestURI struct {
	ProcessId string `uri:"process_id" binding:"required" format:"^p_[0-9a-v]{20}$"`
}
