package primitive

import (
	"github.com/turbot/flowpipe/pipeparser/modconfig"
)

type Primitive interface {
	Run(modconfig.Input) (*modconfig.Output, error)
}
