package execution

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type TriggerExecution struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Args flowpipe.Input `json:"args,omitempty"`
}
