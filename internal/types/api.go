package types

import localconstants "github.com/turbot/flowpipe/internal/constants"

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

type IntegrationRequestURI struct {
	IntegrationName string `uri:"integration_name" binding:"required" format:"^[a-z_]{0,32}$"`
}

type NotifierRequestURI struct {
	NotifierName string `uri:"notifier_name" binding:"required" format:"^[a-z_]{0,32}$"`
}

type TriggerRequestURI struct {
	TriggerName string `uri:"trigger_name" binding:"required" format:"^[a-z]{0,32}$"`
}

type VariableRequestURI struct {
	VariableName string `uri:"variable_name" binding:"required" format:"^[a-z]{0,32}$"`
}

type ModRequestURI struct {
	ModName string `uri:"mod_name" binding:"required" format:"^[a-z]{0,32}$"`
}

type ProcessRequestURI struct {
	// TODO: do we want to pass the ExecutionID or PipelineExecutionID? The log is stored under ExecutionID but the execution works with PipelineExecutionID
	// ProcessId string `uri:"process_id" binding:"required" format:"^(pexec|exec)_[0-9a-v]{20}$"`
	ProcessId string `uri:"process_id" binding:"required" format:"^exec_[0-9a-v]{20}$"`
}

type WebhookRequestUri struct {
	Hook string `json:"hook" uri:"hook" binding:"required"`
	Hash string `json:"hash" uri:"hash" binding:"required"`
}

type WebhookRequestQuery struct {
	WaitTime *int `json:"wait_time" form:"wait_time" binding:"omitempty"`
}

func (c *WebhookRequestQuery) GetWaitTime() int {
	if c.WaitTime != nil {
		return *c.WaitTime
	}
	return localconstants.DefaultWaitRetry
}

type PipelineRequestQuery struct {
	ExecutionMode *string `json:"execution_mode" form:"execution_mode" binding:"omitempty,oneof=synchronous asynchronous"`
}

type InputIDHash struct {
	ID   string `json:"id" uri:"id" binding:"required"`
	Hash string `json:"hash" uri:"hash" binding:"required"`
}
