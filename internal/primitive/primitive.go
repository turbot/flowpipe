package primitive

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

type Primitive interface {
	Run(modconfig.Input) (*modconfig.Output, error)
}
