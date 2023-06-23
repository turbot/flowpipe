package primitive

import "github.com/turbot/flowpipe/internal/types"

type Primitive interface {
	Run(types.Input) (*types.StepOutput, error)
}
