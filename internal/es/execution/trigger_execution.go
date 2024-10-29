package execution

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type TriggerExecution struct {
	ID   string         `json:"id"`
	Name string          `json:"name"`
	Args resources.Input `json:"args,omitempty"`
}
