package types

type Stack struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	StepStatus map[int]string    `json:"pipeline_step_status"`
	Stacks     map[string]*Stack `json:"children"`
}
