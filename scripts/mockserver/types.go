package main

type RoutedInputResponse struct {
	// TODO: #refactor - can we use a shared struct with pipes for this?
	ID              string         `json:"id"`
	TenantID        string         `json:"tenant_id"`
	IdentityID      string         `json:"identity_id"`
	WorkspaceID     string         `json:"workspace_id"`
	NotifierID      string         `json:"notifier_id"`
	Notifier        map[string]any `json:"notifier"`
	ProcessID       string         `json:"process_id"`
	StepExecutionID string         `json:"step_execution_id"`
	RandomID        string         `json:"random_id"`
	State           string         `json:"state"`
	StateReason     string         `json:"state_reason"`
	Inputs          map[string]any `json:"inputs"`
}
