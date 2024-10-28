package primitive

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type Primitive interface {
	Run(flowpipe.Input) (*flowpipe.Output, error)
}
