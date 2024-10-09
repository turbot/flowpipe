package execution

import "github.com/turbot/pipe-fittings/modconfig"

type TriggerExecution struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args modconfig.Input `json:"args,omitempty"`
}
