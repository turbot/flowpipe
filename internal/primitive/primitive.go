package primitive

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type Primitive interface {
	Run(resources.Input) (*resources.Output, error)
}
