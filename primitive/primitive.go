package primitive

import "github.com/turbot/flowpipe/types"

type Primitive interface {
	Run(types.Input) (*types.Output, error)
}
